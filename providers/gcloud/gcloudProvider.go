package gcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	// diff1
	oldRRsMap   map[string]map[key]*gdns.ResourceRecordSet
	zoneNameMap map[string]string
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
	var hc *http.Client
	if key, ok := cfg["private_key"]; ok {
		cfg["private_key"] = strings.Replace(key, "\\n", "\n", -1)
		raw, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		config, err := gauth.JWTConfigFromJSON(raw, "https://www.googleapis.com/auth/ndev.clouddns.readwrite")
		if err != nil {
			return nil, err
		}
		hc = config.Client(ctx)
	} else {
		var err error
		hc, err = gauth.DefaultClient(ctx, "https://www.googleapis.com/auth/ndev.clouddns.readwrite")
		if err != nil {
			return nil, fmt.Errorf("no creds.json private_key found and ADC failed with:\n%s", err)
		}
	}
	// FIXME(tlim): Is it a problem that ctx is included with hc and in
	// the call to NewService?  Seems redundant.
	dcli, err := gdns.NewService(ctx, option.WithHTTPClient(hc))
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
	return g, g.loadZoneInfo()
}

func (g *gcloudProvider) loadZoneInfo() error {
	if g.zones != nil {
		return nil
	}
	g.zones = map[string]*gdns.ManagedZone{}
	pageToken := ""
	for {
		resp, err := g.client.ManagedZones.List(g.project).PageToken(pageToken).Do()
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

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (g *gcloudProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, error) {

	txtutil.SplitSingleLongTxt(dc.Records) // Autosplit long TXT records

	oldRRs, ok := g.oldRRsMap[dc.Name]
	if !ok {
		return nil, fmt.Errorf("oldRRsMap: no zone named %q", dc.Name)
	}
	zoneName, ok := g.zoneNameMap[dc.Name]
	if !ok {
		return nil, fmt.Errorf("zoneNameMap: no zone named %q", dc.Name)
	}

	// first collect keys that have changed
	var differ diff.Differ
	if !diff2.EnableDiff2 {
		differ = diff.New(dc)
	} else {
		differ = diff.NewCompat(dc)
	}
	_, create, delete, modify, err := differ.IncrementalDiff(existingRecords)
	if err != nil {
		return nil, fmt.Errorf("incdiff error: %w", err)
	}

	changedKeys := map[key]bool{}
	var msgs []string
	for _, c := range create {
		msgs = append(msgs, fmt.Sprint(c))
		changedKeys[keyForRec(c.Desired)] = true
	}
	for _, d := range delete {
		msgs = append(msgs, fmt.Sprint(d))
		changedKeys[keyForRec(d.Existing)] = true
	}
	for _, m := range modify {
		msgs = append(msgs, fmt.Sprint(m))
		changedKeys[keyForRec(m.Existing)] = true
	}
	if len(changedKeys) == 0 {
		return nil, nil
	}
	chg := &gdns.Change{Kind: "dns#change"}
	for ck := range changedKeys {
		// remove old version (if present)
		if old, ok := oldRRs[ck]; ok {
			chg.Deletions = append(chg.Deletions, old)
		}
		// collect records to replace with
		newRRs := &gdns.ResourceRecordSet{
			Name: ck.Name,
			Type: ck.Type,
			Kind: "dns#resourceRecordSet",
		}
		for _, r := range dc.Records {
			if keyForRec(r) == ck {
				newRRs.Rrdatas = append(newRRs.Rrdatas, r.GetTargetCombined())
				newRRs.Ttl = int64(r.TTL)
			}
		}
		if len(newRRs.Rrdatas) > 0 {
			chg.Additions = append(chg.Additions, newRRs)
		}
	}

	// FIXME(tlim): Google will return an error if too many changes are
	// specified in a single request. We should split up very large
	// batches.  This can be reliably reproduced with the 1201
	// integration test.  The error you get is:
	// googleapi: Error 403: The change would exceed quota for additions per change., quotaExceeded
	//log.Printf("PAUSE STT = %+v %v\n", err, resp)
	//log.Printf("PAUSE ERR = %+v %v\n", err, resp)

	runChange := func() error {
	retry:
		resp, err := g.client.Changes.Create(g.project, zoneName, chg).Do()
		if retryNeeded(resp, err) {
			goto retry
		}
		if err != nil {
			return fmt.Errorf("runChange error: %w", err)
		}
		return nil
	}

	return []*models.Correction{{
		Msg: strings.Join(msgs, "\n"),
		F:   runChange,
	}}, nil

}

func nativeToRecord(set *gdns.ResourceRecordSet, rec, origin string) (*models.RecordConfig, error) {
	r := &models.RecordConfig{}
	r.SetLabelFromFQDN(set.Name, origin)
	r.TTL = uint32(set.Ttl)
	if err := r.PopulateFromString(set.Type, rec, origin); err != nil {
		return nil, fmt.Errorf("unparsable record received from GCLOUD: %w", err)
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
		resp, err := call.Do()
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
	if g.nameServerSet != nil {
		printer.Printf("Adding zone for %s to gcloud account with name_server_set %s\n", domain, *g.nameServerSet)
		mz = &gdns.ManagedZone{
			DnsName:       domain + ".",
			NameServerSet: *g.nameServerSet,
			Name:          "zone-" + strings.Replace(domain, ".", "-", -1),
			Description:   "zone added by dnscontrol",
		}
	} else {
		printer.Printf("Adding zone for %s to gcloud account \n", domain)
		mz = &gdns.ManagedZone{
			DnsName:     domain + ".",
			Name:        "zone-" + strings.Replace(domain, ".", "-", -1),
			Description: "zone added by dnscontrol",
		}
	}
	g.zones = nil // reset cache
	_, err = g.client.ManagedZones.Create(g.project, mz).Do()
	return err
}

const initialBackoff = time.Second * 10 // First delay duration
const maxBackoff = time.Minute * 3      // Maximum backoff delay

// backoff is the amount of time to sleep if a 429 or 504 is received.
// It is doubled after each use.
var backoff = initialBackoff
var backoff404 = false // Set if the last call requested a retry of a 404

func retryNeeded(resp *gdns.Change, err error) bool {
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
