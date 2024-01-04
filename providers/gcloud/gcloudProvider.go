package gcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/StackExchange/dnscontrol/v4/pkg/txtutil"
	"github.com/StackExchange/dnscontrol/v4/providers"
	gauth "golang.org/x/oauth2/google"
	gdns "google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const selfLinkBasePath = "https://www.googleapis.com/compute/v1/projects/"

var features = providers.DocumentationNotes{
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDSForChildren:    providers.Can(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Can(),
}

var (
	visibilityCheck  = regexp.MustCompile("^(public|private)$")
	networkURLCheck  = regexp.MustCompile("^" + selfLinkBasePath + "[a-z][-a-z0-9]{4,28}[a-z0-9]/global/networks/[a-z]([-a-z0-9]{0,61}[a-z0-9])?$")
	networkNameCheck = regexp.MustCompile("^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$")
)

func sPtr(s string) *string {
	return &s
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   New,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("GCLOUD", fns, features)
}

type gcloudProvider struct {
	client        *gdns.Service
	project       string
	nameServerSet *string
	zones         map[string]*gdns.ManagedZone
	// For use with diff / NewComnpat()
	oldRRsMap   map[string]map[key]*gdns.ResourceRecordSet
	zoneNameMap map[string]string
	// provider metadata fields
	Visibility string   `json:"visibility"`
	Networks   []string `json:"networks"`
}

type errNoExist struct {
	domain string
}

func (e errNoExist) Error() string {
	return fmt.Sprintf("Domain '%s' not found in gcloud account", e.domain)
}

// New creates a new gcloud provider
func New(cfg map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	// the key as downloaded is json encoded with literal "\n" instead of newlines.
	// in some cases (round-tripping through env vars) this tends to get messed up.
	// fix it if we find that.

	ctx := context.Background()
	var opt option.ClientOption
	if key, ok := cfg["private_key"]; ok {
		cfg["private_key"] = strings.Replace(key, "\\n", "\n", -1)
		raw, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		config, err := gauth.JWTConfigFromJSON(raw, gdns.NdevClouddnsReadwriteScope)
		if err != nil {
			return nil, err
		}
		opt = option.WithTokenSource(config.TokenSource(ctx))
	} else {
		opt = option.WithScopes(gdns.NdevClouddnsReadwriteScope)
	}
	dcli, err := gdns.NewService(ctx, opt)
	if err != nil {
		return nil, err
	}
	var nss *string
	if val, ok := cfg["name_server_set"]; ok {
		printer.Printf("GCLOUD :name_server_set %s configured\n", val)
		nss = sPtr(val)
	}

	g := &gcloudProvider{
		client:        dcli,
		nameServerSet: nss,
		project:       cfg["project_id"],
		oldRRsMap:     map[string]map[key]*gdns.ResourceRecordSet{},
		zoneNameMap:   map[string]string{},
	}
	if len(metadata) != 0 {
		err := json.Unmarshal(metadata, g)
		if err != nil {
			return nil, err
		}
		if len(g.Visibility) != 0 {
			if ok := visibilityCheck.MatchString(g.Visibility); !ok {
				return nil, fmt.Errorf("GCLOUD :visibility set but not one of \"public\" or \"private\"")
			}
			printer.Printf("GCLOUD :visibility %s configured\n", g.Visibility)
		}
		for i, v := range g.Networks {
			if ok := networkURLCheck.MatchString(v); ok {
				// the user specified a fully qualified network url
				continue
			}
			if ok := networkNameCheck.MatchString(v); !ok {
				return nil, fmt.Errorf("GCLOUD :networks set but %s does not appear to be a valid network name or url", v)
			}
			// assume target vpc network exists in the same project as the dns zones
			g.Networks[i] = fmt.Sprintf("%s%s/global/networks/%s", selfLinkBasePath, g.project, v)
		}
	}
	return g, g.loadZoneInfo()
}

func (g *gcloudProvider) loadZoneInfo() error {
	// TODO(asn-iac): In order to fully support split horizon domains within the same GCP project,
	// need to parse the zone Visibility field from *ManagedZone, but currently
	// gcloudProvider.zones is map[string]*gdns.ManagedZone
	// where the map keys are the zone dns names. A given GCP project can have
	// multiple zones of the same dns name.
	if g.zones != nil {
		return nil
	}
	g.zones = map[string]*gdns.ManagedZone{}
	pageToken := ""
	for {
	retry:
		resp, err := g.client.ManagedZones.List(g.project).PageToken(pageToken).Do()
		var check *googleapi.ServerResponse
		if resp != nil {
			check = &resp.ServerResponse
		}
		if retryNeeded(check, err) {
			goto retry
		}
		if err != nil {
			return err
		}
		for _, z := range resp.ManagedZones {
			g.zones[z.DnsName] = z
		}
		if pageToken = resp.NextPageToken; pageToken == "" {
			break
		}
	}
	return nil
}

// ListZones returns the list of zones (domains) in this account.
func (g *gcloudProvider) ListZones() ([]string, error) {
	var zones []string
	for i := range g.zones {
		zones = append(zones, strings.TrimSuffix(i, "."))
	}
	return zones, nil
}

func (g *gcloudProvider) getZone(domain string) (*gdns.ManagedZone, error) {
	return g.zones[domain+"."], nil
}

func (g *gcloudProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	zone, err := g.getZone(domain)
	if err != nil {
		return nil, err
	}
	if zone == nil {
		return nil, fmt.Errorf("domain %q not found in your GCLOUD account", domain)
	}
	return models.ToNameserversStripTD(zone.NameServers)
}

type key struct {
	Type string
	Name string
}

func keyFor(r *gdns.ResourceRecordSet) key {
	return key{Type: r.Type, Name: r.Name}
}
func keyForRec(r *models.RecordConfig) key {
	return key{Type: r.Type, Name: r.GetLabelFQDN() + "."}
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (g *gcloudProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	existingRecords, err := g.getZoneSets(domain)
	return existingRecords, err
}

func (g *gcloudProvider) getZoneSets(domain string) (models.Records, error) {
	rrs, zoneName, err := g.getRecords(domain)
	if err != nil {
		return nil, err
	}
	// convert to dnscontrol RecordConfig format
	existingRecords := []*models.RecordConfig{}
	oldRRs := map[key]*gdns.ResourceRecordSet{}
	for _, set := range rrs {
		oldRRs[keyFor(set)] = set
		for _, rec := range set.Rrdatas {
			rt, err := nativeToRecord(set, rec, domain)
			if err != nil {
				return nil, err
			}

			existingRecords = append(existingRecords, rt)
		}
	}

	g.oldRRsMap[domain] = oldRRs
	g.zoneNameMap[domain] = zoneName

	return existingRecords, err
}

type msgs struct {
	Additions, Deletions []string
}

type orderedChanges struct {
	Change *gdns.Change
	Msgs   msgs
}

type correctionValues struct {
	Change *gdns.Change
	Msgs   string
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (g *gcloudProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, error) {

	changes, err := diff2.ByRecordSet(existingRecords, dc, nil)
	if err != nil {
		return nil, err
	}
	if len(changes) == 0 {
		return nil, nil
	}

	var corrections []*models.Correction
	batch := &gdns.Change{Kind: "dns#change"}
	var accumlatedMsgs []string
	var newMsgs []string
	var newAdds, newDels *gdns.ResourceRecordSet

	for _, change := range changes {
		n := change.Key.NameFQDN + "."
		ty := change.Key.Type
		newMsgs = nil
		if len(change.Msgs) != 0 {
			newMsgs = change.Msgs
		}
		switch change.Type {
		case diff2.REPORT:
			newAdds = nil
			newDels = nil
		case diff2.CREATE:
			newAdds = mkRRSs(n, ty, change.New)
			newDels = nil
		case diff2.CHANGE:
			newAdds = mkRRSs(n, ty, change.New)
			newDels = mkRRSs(n, ty, change.Old)
		case diff2.DELETE:
			newAdds = nil
			newDels = mkRRSs(n, ty, change.Old)
		default:
			panic(fmt.Sprintf("unhandled change.TYPE %s", change.Type))
		}

		// If the work would overflow the current batch, process what we have so far and start a new batch.
		if wouldOverfill(batch, newAdds, newDels) {
			corrections = g.mkCorrection(corrections, accumlatedMsgs, batch, dc.Name)

			// Start a new batch.
			batch = &gdns.Change{Kind: "dns#change"}
			accumlatedMsgs = nil
		}

		// Add the new work to the batch.
		if newAdds != nil {
			batch.Additions = append(batch.Additions, newAdds)
		}
		if newDels != nil {
			batch.Deletions = append(batch.Deletions, newDels)
		}
		if len(accumlatedMsgs) != 0 {
			accumlatedMsgs = append(accumlatedMsgs, newMsgs...)
		}

	}

	corrections = g.mkCorrection(corrections, accumlatedMsgs, batch, dc.Name)

	return corrections, nil
}

func (g *gcloudProvider) mkCorrection(corrections []*models.Correction, accumulatedMsgs []string, batch *gdns.Change, origin string) []*models.Correction {
	if len(accumulatedMsgs) == 0 && len(batch.Additions) == 0 && len(batch.Deletions) == 0 {
		// Nothing to do!
		fmt.Fprintf(os.Stdout, "DEBUG: nothing to do!\n")
		return corrections
	}

	corr := &models.Correction{}
	if len(accumulatedMsgs) != 0 {
		corr.Msg = strings.Join(accumulatedMsgs, "\n")
		fmt.Fprintf(os.Stdout, "DEBUG: msgs added msg=%v\n", accumulatedMsgs)
	}
	if len(batch.Additions) != 0 || len(batch.Deletions) == 0 {
		fmt.Fprintf(os.Stdout, "DEBUG: adds=%d dels=%d\n", len(batch.Additions), len(batch.Deletions))
		fmt.Fprintf(os.Stdout, "DEBUG: adds=%v\n", batch.Additions)
		fmt.Fprintf(os.Stdout, "DEBUG: dels=%v\n", batch.Deletions)
		// Only set "F" if there is work to do. F = nil tells the caller this is a "message", not an action.
		corr.F = func() error { return g.process(origin, batch) }
	}

	// corrections = append(corrections, corr)
	// return corrections
	return append(corrections, corr)
}

// mkRRSs returns a gdns.ResourceRecordSet using the name, rType, and recs
func mkRRSs(name, rType string, recs models.Records) *gdns.ResourceRecordSet {
	newRRS := &gdns.ResourceRecordSet{
		Name: name,
		Type: rType,
		Kind: "dns#resourceRecordSet",
	}

	newRRS.Ttl = int64(recs[0].TTL) // Assume all TTLs are the same. diff2 assures they are.
	for _, r := range recs {
		newRRS.Rrdatas = append(newRRS.Rrdatas, r.GetTargetCombinedFunc(txtutil.EncodeQuoted))

		// Test that assumption that diff2 assures all TTLs in a recordset are the same.
		if newRRS.Ttl != int64(r.TTL) {
			panic("TTLs not the same")
		}

	}

	return newRRS
}

func wouldOverfill(batch *gdns.Change, adds, dels *gdns.ResourceRecordSet) bool {
	const batchMax = 1000

	addCount := 0
	if adds != nil {
		addCount = len(adds.Rrdatas)
	}
	delCount := 0
	if dels != nil {
		delCount = len(dels.Rrdatas)
	}

	if (len(batch.Additions) + addCount) > batchMax {
		return true
	}
	if (len(batch.Deletions) + delCount) > batchMax {
		return true
	}
	return false
}

func (g *gcloudProvider) process(origin string, batch *gdns.Change) error {

	zoneName, ok := g.zoneNameMap[origin]
	if !ok {
		return fmt.Errorf("zoneNameMap: no zone named %q", origin)
	}

retry:
	resp, err := g.client.Changes.Create(g.project, zoneName, batch).Do()
	var check *googleapi.ServerResponse
	if resp != nil {
		check = &resp.ServerResponse
	}
	if retryNeeded(check, err) {
		goto retry
	}
	if err != nil {
		return fmt.Errorf("runChange error: %w", err)
	}
	return nil
}

// OLDGetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (g *gcloudProvider) OLDGetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, error) {

	oldRRs, ok := g.oldRRsMap[dc.Name]
	if !ok {
		return nil, fmt.Errorf("oldRRsMap: no zone named %q", dc.Name)
	}
	zoneName, ok := g.zoneNameMap[dc.Name]
	if !ok {
		return nil, fmt.Errorf("zoneNameMap: no zone named %q", dc.Name)
	}

	// first collect keys that have changed
	toReport, create, toDelete, modify, err := diff.NewCompat(dc).IncrementalDiff(existingRecords)
	if err != nil {
		return nil, fmt.Errorf("incdiff error: %w", err)
	}
	// Start corrections with the reports
	corrections := diff.GenerateMessageCorrections(toReport)

	// Now generate all other corrections

	changedKeys := map[key]string{}
	for _, c := range create {
		msg := fmt.Sprintln(c)
		if k, ok := changedKeys[keyForRec(c.Desired)]; ok {
			msg = strings.Join([]string{k, msg}, "")
		}
		changedKeys[keyForRec(c.Desired)] = msg
	}
	for _, d := range toDelete {
		msg := fmt.Sprintln(d)
		if k, ok := changedKeys[keyForRec(d.Existing)]; ok {
			msg = strings.Join([]string{k, msg}, "")
		}
		changedKeys[keyForRec(d.Existing)] = msg
	}
	for _, m := range modify {
		msg := fmt.Sprintln(m)
		if k, ok := changedKeys[keyForRec(m.Existing)]; ok {
			msg = strings.Join([]string{k, msg}, "")
		}
		changedKeys[keyForRec(m.Existing)] = msg
	}
	if len(changedKeys) == 0 {
		return nil, nil
	}
	chg := orderedChanges{Change: &gdns.Change{}, Msgs: msgs{}}
	// create slices of Deletions and Additions
	// that can be split into properly ordered batches
	// if necessary.  Retain the string messages from
	// differ in the same order
	for ck, msg := range changedKeys {
		newRRs := &gdns.ResourceRecordSet{
			Name: ck.Name,
			Type: ck.Type,
			Kind: "dns#resourceRecordSet",
		}
		for _, r := range dc.Records {
			if keyForRec(r) == ck {
				newRRs.Rrdatas = append(newRRs.Rrdatas, r.GetTargetCombinedFunc(txtutil.EncodeQuoted))
				newRRs.Ttl = int64(r.TTL)
			}
		}
		if len(newRRs.Rrdatas) > 0 {
			// if we have Rrdatas because the key from differ
			// exists in normalized config,
			// check whether the key also has data in oldRRs.
			// if so, this is actually a modify operation, insert
			// the Addition and Deletion at the beginning of the slices
			// to ensure they are executed in the same batch
			if old, ok := oldRRs[ck]; ok {
				chg.Change.Additions = append([]*gdns.ResourceRecordSet{newRRs}, chg.Change.Additions...)
				chg.Change.Deletions = append([]*gdns.ResourceRecordSet{old}, chg.Change.Deletions...)
				chg.Msgs.Additions = append([]string{msg}, chg.Msgs.Additions...)
				chg.Msgs.Deletions = append([]string{""}, chg.Msgs.Deletions...)
			} else {
				// otherwise this is a pure Addition
				chg.Change.Additions = append(chg.Change.Additions, newRRs)
				chg.Msgs.Additions = append(chg.Msgs.Additions, msg)
			}
		} else {
			// there is no Rrdatas from normalized config for this key.
			// it must be a Deletion, use the ResourceRecordSet from
			// oldRRs
			if old, ok := oldRRs[ck]; ok {
				chg.Change.Deletions = append(chg.Change.Deletions, old)
				chg.Msgs.Deletions = append(chg.Msgs.Deletions, msg)
			}
		}
	}

	// create a slice of Changes in batches of at most
	// 1000 Deletions and 1000 Additions per Change.
	// create a slice of strings that aligns with the batch
	// to output with each correction/Change
	const batchMax = 1000
	setBatchLen := func(len int) int {
		if len > batchMax {
			return batchMax
		}
		return len
	}
	chgSet := []correctionValues{}
	for len(chg.Change.Deletions) > 0 {
		b := setBatchLen(len(chg.Change.Deletions))
		chgSet = append(chgSet, correctionValues{Change: &gdns.Change{Deletions: chg.Change.Deletions[:b:b], Kind: "dns#change"}, Msgs: strings.Join(chg.Msgs.Deletions[:b:b], "")})
		chg.Change.Deletions = chg.Change.Deletions[b:]
		chg.Msgs.Deletions = chg.Msgs.Deletions[b:]
	}
	for i := 0; len(chg.Change.Additions) > 0; i++ {
		b := setBatchLen(len(chg.Change.Additions))
		if len(chgSet) == i {
			chgSet = append(chgSet, correctionValues{Change: &gdns.Change{Additions: chg.Change.Additions[:b:b], Kind: "dns#change"}, Msgs: strings.Join(chg.Msgs.Additions[:b:b], "")})
		} else {
			chgSet[i].Change.Additions = chg.Change.Additions[:b:b]
			chgSet[i].Msgs += strings.Join(chg.Msgs.Additions[:b:b], "")
		}
		chg.Change.Additions = chg.Change.Additions[b:]
		chg.Msgs.Additions = chg.Msgs.Additions[b:]
	}
	// create a Correction for each gdns.Change
	// that needs to be executed
	makeCorrection := func(chg *gdns.Change, msgs string) {
		runChange := func() error {
		retry:
			resp, err := g.client.Changes.Create(g.project, zoneName, chg).Do()
			var check *googleapi.ServerResponse
			if resp != nil {
				check = &resp.ServerResponse
			}
			if retryNeeded(check, err) {
				goto retry
			}
			if err != nil {
				return fmt.Errorf("runChange error: %w", err)
			}
			return nil
		}
		corrections = append(corrections,
			&models.Correction{
				Msg: strings.TrimSuffix(msgs, "\n"),
				F:   runChange,
			})
	}
	for _, v := range chgSet {
		makeCorrection(v.Change, v.Msgs)
	}

	return corrections, nil
}

func nativeToRecord(set *gdns.ResourceRecordSet, rec, origin string) (*models.RecordConfig, error) {
	r := &models.RecordConfig{}
	r.SetLabelFromFQDN(set.Name, origin)
	r.TTL = uint32(set.Ttl)
	rtype := set.Type
	err := r.PopulateFromStringFunc(rtype, rec, origin, txtutil.ParseQuoted)
	if err != nil {
		return nil, fmt.Errorf("unparsable record %q received from GCLOUD: %w", rtype, err)
	}
	return r, nil
}

func (g *gcloudProvider) getRecords(domain string) ([]*gdns.ResourceRecordSet, string, error) {
	zone, err := g.getZone(domain)
	if err != nil {
		return nil, "", err
	}
	pageToken := ""
	sets := []*gdns.ResourceRecordSet{}
	for {
		call := g.client.ResourceRecordSets.List(g.project, zone.Name)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
	retry:
		resp, err := call.Do()
		var check *googleapi.ServerResponse
		if resp != nil {
			check = &resp.ServerResponse
		}
		if retryNeeded(check, err) {
			goto retry
		}
		if err != nil {
			return nil, "", err
		}
		for _, rrs := range resp.Rrsets {
			if rrs.Type == "SOA" {
				continue
			}
			sets = append(sets, rrs)
		}
		if pageToken = resp.NextPageToken; pageToken == "" {
			break
		}
	}
	return sets, zone.Name, nil
}

func (g *gcloudProvider) EnsureZoneExists(domain string) error {
	z, err := g.getZone(domain)
	if err != nil {
		if _, ok := err.(errNoExist); !ok {
			return err
		}
	}
	if z != nil {
		return nil
	}
	var mz *gdns.ManagedZone
	printer.Printf("Adding zone for %s to gcloud account ", domain)
	mz = &gdns.ManagedZone{
		DnsName:     domain + ".",
		Name:        "zone-" + strings.Replace(domain, ".", "-", -1),
		Description: "zone added by dnscontrol",
	}
	if g.nameServerSet != nil {
		mz.NameServerSet = *g.nameServerSet
		printer.Printf("with name_server_set %s ", *g.nameServerSet)
	}
	if len(g.Visibility) != 0 {
		mz.Visibility = g.Visibility
		printer.Printf("with %s visibility ", g.Visibility)
		// prevent possible GCP resource name conflicts when split horizon can be properly implemented
		mz.Name = strings.Replace(mz.Name, "zone-", "zone-"+g.Visibility+"-", 1)
	}
	if g.Networks != nil {
		mzn := make([]*gdns.ManagedZonePrivateVisibilityConfigNetwork, len(g.Networks))
		printer.Printf("for network(s) ")
		for _, v := range g.Networks {
			printer.Printf("%s ", v)
			mzn = append(mzn, &gdns.ManagedZonePrivateVisibilityConfigNetwork{NetworkUrl: v})
		}
		mz.PrivateVisibilityConfig = &gdns.ManagedZonePrivateVisibilityConfig{Networks: mzn}
	}
	printer.Printf("\n")
	g.zones[domain+"."], err = g.client.ManagedZones.Create(g.project, mz).Do()
	return err
}

const initialBackoff = time.Second * 10 // First delay duration
const maxBackoff = time.Minute * 3      // Maximum backoff delay

// backoff is the amount of time to sleep if a 429 or 504 is received.
// It is doubled after each use.
var backoff = initialBackoff
var backoff404 = false // Set if the last call requested a retry of a 404

func retryNeeded(resp *googleapi.ServerResponse, err error) bool {
	if err != nil {
		return false // Not an error.
	}
	serr, ok := err.(*googleapi.Error)
	if !ok {
		return false // Not a google error.
	}
	if serr.Code == 200 {
		backoff = initialBackoff // Reset
		return false             // Success! No need to retry.
	}

	if serr.Code == 404 {
		// serr.Code == 404 happens occasionally when GCLOUD hasn't
		// finished updating the database yet.  We pause and retry
		// exactly once. There should be a better way to do this, such as
		// a callback that would tell us a transaction is complete.
		if backoff404 {
			backoff404 = false
			return false // Give up. We've done this already.
		}
		log.Printf("Special 404 pause-and-retry for GCLOUD: Pausing %s\n", backoff)
		time.Sleep(backoff)
		backoff404 = true
		return true // Request a retry.
	}
	backoff404 = false

	if serr.Code != 429 && serr.Code != 502 && serr.Code != 503 {
		return false // Not an error that permits retrying.
	}

	// TODO(tlim): In theory, resp.Header has a header that says how
	// long to wait but I haven't been able to capture that header in
	// the wild. If you get these "RUNCHANGE HEAD" messages, please
	// file a bug with the contents!

	if resp != nil {
		log.Printf("NOTE: If you see this message, please file a bug with the output below:\n")
		log.Printf("RUNCHANGE CODE = %+v\n", resp.HTTPStatusCode)
		log.Printf("RUNCHANGE HEAD = %+v\n", resp.Header)
	}

	// a simple exponential back-off
	log.Printf("Pausing due to ratelimit: %v seconds\n", backoff)
	time.Sleep(backoff)
	backoff = backoff + (backoff / 2)
	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	return true // Request the API call be re-tried.
}
