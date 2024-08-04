package gcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/v4/models"
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
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Can(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDSForChildren:    providers.Can(),
	providers.CanUseHTTPS:            providers.Can(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseSVCB:             providers.Can(),
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
	const providerName = "GCLOUD"
	const providerMaintainer = "@riyadhalnur"
	fns := providers.DspFuncs{
		Initializer:   New,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	providers.RegisterMaintainer(providerName, providerMaintainer)
}

type gcloudProvider struct {
	client        *gdns.Service
	project       string
	nameServerSet *string
	zones         map[string]*gdns.ManagedZone
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

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (g *gcloudProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	existingRecords, err := g.getZoneSets(domain)
	return existingRecords, err
}

func (g *gcloudProvider) getZoneSets(domain string) (models.Records, error) {
	rrs, err := g.getRecords(domain)
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

	return existingRecords, err
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

		// Determine the work to be done.
		n := change.Key.NameFQDN + "."
		ty := change.Key.Type
		switch change.Type {
		case diff2.REPORT:
			newMsgs = change.Msgs
			newAdds = nil
			newDels = nil
		case diff2.CREATE:
			newMsgs = change.Msgs
			newAdds = mkRRSs(n, ty, change.New)
			newDels = nil
		case diff2.CHANGE:
			newMsgs = change.Msgs
			newAdds = mkRRSs(n, ty, change.New)
			newDels = change.Old[0].Original.(*gdns.ResourceRecordSet)
		case diff2.DELETE:
			newMsgs = change.Msgs
			newAdds = nil
			newDels = change.Old[0].Original.(*gdns.ResourceRecordSet)
		default:
			return nil, fmt.Errorf("GCLOUD unhandled change.TYPE %s", change.Type)
		}

		// If the work would overflow the current batch, process what we have so far and start a new batch.
		if wouldOverfill(batch, newAdds, newDels) {
			// Process what we have.
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
		if len(newMsgs) != 0 {
			accumlatedMsgs = append(accumlatedMsgs, newMsgs...)
		}

	}

	// Process the remaining work.
	corrections = g.mkCorrection(corrections, accumlatedMsgs, batch, dc.Name)
	return corrections, nil
}

// mkRRSs returns a gdns.ResourceRecordSet using the name, rType, and recs
func mkRRSs(name, rType string, recs models.Records) *gdns.ResourceRecordSet {
	if len(recs) == 0 { // NB(tlim): This is defensive. mkRRSs is never called with an empty list.
		return nil
	}

	newRRS := &gdns.ResourceRecordSet{
		Name: name,
		Type: rType,
		Kind: "dns#resourceRecordSet",
		Ttl:  int64(recs[0].TTL), // diff2 assures all TTLs in a ReceordSet are the same.
	}

	for _, r := range recs {
		newRRS.Rrdatas = append(newRRS.Rrdatas, r.GetTargetCombinedFunc(txtutil.EncodeQuoted))
	}

	return newRRS
}

// wouldOverfill returns true if adding this work would overflow the batch.
func wouldOverfill(batch *gdns.Change, adds, dels *gdns.ResourceRecordSet) bool {
	const batchMax = 1000
	// Google used to document batchMax = 1000.  As of 2024-01 the max isn't
	// documented but testing shows it rejects if either Additions or Deletions
	// are >3000.  Setting this to 3001 makes the batchRecordswithOthers
	// integration test fail.
	// It is currently set to 1000 because (1) its the last documented max,
	// (2) changes of more than 1000 RSets is rare; we'd rather be correct and
	// working than broken and efficient.

	addCount := 0
	if adds != nil {
		addCount = len(adds.Rrdatas)
	}
	delCount := 0
	if dels != nil {
		delCount = len(dels.Rrdatas)
	}

	if (len(batch.Additions) + addCount) > batchMax { // Would additions push us over the limit?
		return true
	}
	if (len(batch.Deletions) + delCount) > batchMax { // Would deletions push us over the limit?
		return true
	}
	return false
}

func (g *gcloudProvider) mkCorrection(corrections []*models.Correction, accumulatedMsgs []string, batch *gdns.Change, origin string) []*models.Correction {
	if len(accumulatedMsgs) == 0 && len(batch.Additions) == 0 && len(batch.Deletions) == 0 {
		// Nothing to do!
		return corrections
	}

	corr := &models.Correction{}
	if len(accumulatedMsgs) != 0 {
		corr.Msg = strings.Join(accumulatedMsgs, "\n")
	}
	if (len(batch.Additions) + len(batch.Deletions)) != 0 {
		corr.F = func() error { return g.process(origin, batch) }
	}

	corrections = append(corrections, corr)
	return corrections
}

// process calls the Google DNS API to process a Change and re-tries if needed.
func (g *gcloudProvider) process(origin string, batch *gdns.Change) error {

	zoneName, err := g.getZone(origin)
	if err != nil || zoneName == nil {
		return fmt.Errorf("zoneNameMap: no zone named %q", origin)
	}

retry:
	resp, err := g.client.Changes.Create(g.project, zoneName.Name, batch).Do()
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

func nativeToRecord(set *gdns.ResourceRecordSet, rec, origin string) (*models.RecordConfig, error) {
	r := &models.RecordConfig{}
	r.SetLabelFromFQDN(set.Name, origin)
	r.TTL = uint32(set.Ttl)
	rtype := set.Type
	r.Original = set
	err := r.PopulateFromStringFunc(rtype, rec, origin, txtutil.ParseQuoted)
	if err != nil {
		return nil, fmt.Errorf("unparsable record %q received from GCLOUD: %w", rtype, err)
	}
	return r, nil
}

func (g *gcloudProvider) getRecords(domain string) ([]*gdns.ResourceRecordSet, error) {
	zone, err := g.getZone(domain)
	if err != nil {
		return nil, err
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
			return nil, err
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
	return sets, nil
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
		mzn := make([]*gdns.ManagedZonePrivateVisibilityConfigNetwork, 0, len(g.Networks))
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
