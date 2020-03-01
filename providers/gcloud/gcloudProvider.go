package google

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	gauth "golang.org/x/oauth2/google"
	gdns "google.golang.org/api/dns/v1"

	"github.com/StackExchange/dnscontrol/v2/models"
	"github.com/StackExchange/dnscontrol/v2/providers"
	"github.com/StackExchange/dnscontrol/v2/providers/diff"
)

var features = providers.DocumentationNotes{
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseTXTMulti:         providers.Can(),
	providers.CanGetZones:            providers.Can(),
}

func sPtr(s string) *string {
	return &s
}

func init() {
	providers.RegisterDomainServiceProviderType("GCLOUD", New, features)
}

type gcloud struct {
	client        *gdns.Service
	project       string
	nameServerSet *string
	zones         map[string]*gdns.ManagedZone
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
	if key, ok := cfg["private_key"]; ok {
		cfg["private_key"] = strings.Replace(key, "\\n", "\n", -1)
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	config, err := gauth.JWTConfigFromJSON(raw, "https://www.googleapis.com/auth/ndev.clouddns.readwrite")
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	hc := config.Client(ctx)
	dcli, err := gdns.New(hc)
	if err != nil {
		return nil, err
	}
	var nss *string = nil
	if val, ok := cfg["name_server_set"]; ok {
		fmt.Printf("GCLOUD :name_server_set %s configured\n", val)
		nss = sPtr(val)
	}

	g := &gcloud{
		client:        dcli,
		nameServerSet: nss,
		project:       cfg["project_id"],
	}
	return g, g.loadZoneInfo()
}

func (g *gcloud) loadZoneInfo() error {
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
func (g *gcloud) ListZones() ([]string, error) {
	var zones []string
	for i := range g.zones {
		zones = append(zones, strings.TrimSuffix(i, "."))
	}
	return zones, nil
}

func (g *gcloud) getZone(domain string) (*gdns.ManagedZone, error) {
	return g.zones[domain+"."], nil
}

func (g *gcloud) GetNameservers(domain string) ([]*models.Nameserver, error) {
	zone, err := g.getZone(domain)
	if err != nil {
		return nil, err
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
func (g *gcloud) GetZoneRecords(domain string) (models.Records, error) {
	existingRecords, _, _, err := g.getZoneSets(domain)
	return existingRecords, err
}

func (g *gcloud) getZoneSets(domain string) (models.Records, map[key]*gdns.ResourceRecordSet, string, error) {
	rrs, zoneName, err := g.getRecords(domain)
	if err != nil {
		return nil, nil, "", err
	}
	// convert to dnscontrol RecordConfig format
	existingRecords := []*models.RecordConfig{}
	oldRRs := map[key]*gdns.ResourceRecordSet{}
	for _, set := range rrs {
		oldRRs[keyFor(set)] = set
		for _, rec := range set.Rrdatas {
			existingRecords = append(existingRecords, nativeToRecord(set, rec, domain))
		}
	}
	return existingRecords, oldRRs, zoneName, err
}

func (g *gcloud) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	if err := dc.Punycode(); err != nil {
		return nil, err
	}
	existingRecords, oldRRs, zoneName, err := g.getZoneSets(dc.Name)
	if err != nil {
		return nil, err
	}

	// Normalize
	models.PostProcessRecords(existingRecords)

	// first collect keys that have changed
	differ := diff.New(dc)
	_, create, delete, modify := differ.IncrementalDiff(existingRecords)
	changedKeys := map[key]bool{}
	desc := ""
	for _, c := range create {
		desc += fmt.Sprintln(c)
		changedKeys[keyForRec(c.Desired)] = true
	}
	for _, d := range delete {
		desc += fmt.Sprintln(d)
		changedKeys[keyForRec(d.Existing)] = true
	}
	for _, m := range modify {
		desc += fmt.Sprintln(m)
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

	runChange := func() error {
		_, err := g.client.Changes.Create(g.project, zoneName, chg).Do()
		return err
	}
	return []*models.Correction{{
		Msg: desc,
		F:   runChange,
	}}, nil
}

func nativeToRecord(set *gdns.ResourceRecordSet, rec, origin string) *models.RecordConfig {
	r := &models.RecordConfig{}
	r.SetLabelFromFQDN(set.Name, origin)
	r.TTL = uint32(set.Ttl)
	if err := r.PopulateFromString(set.Type, rec, origin); err != nil {
		panic(fmt.Errorf("unparsable record received from GCLOUD: %w", err))
	}
	return r
}

func (g *gcloud) getRecords(domain string) ([]*gdns.ResourceRecordSet, string, error) {
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

func (g *gcloud) EnsureDomainExists(domain string) error {
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
		fmt.Printf("Adding zone for %s to gcloud account with name_server_set %s\n", domain, *g.nameServerSet)
		mz = &gdns.ManagedZone{
			DnsName:       domain + ".",
			NameServerSet: *g.nameServerSet,
			Name:          "zone-" + strings.Replace(domain, ".", "-", -1),
			Description:   "zone added by dnscontrol",
		}
	} else {
		fmt.Printf("Adding zone for %s to gcloud account \n", domain)
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
