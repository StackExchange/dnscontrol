package google

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	gauth "golang.org/x/oauth2/google"
	gdns "google.golang.org/api/dns/v1"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/pkg/errors"
)

var features = providers.DocumentationNotes{
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseCAA:              providers.Can(),
}

func init() {
	providers.RegisterDomainServiceProviderType("GCLOUD", New, features)
}

type gcloud struct {
	client  *gdns.Service
	project string
	zones   map[string]*gdns.ManagedZone
}

// New creates a new gcloud provider
func New(cfg map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
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
	return &gcloud{
		client:  dcli,
		project: cfg["project_id"],
	}, nil
}

type errNoExist struct {
	domain string
}

func (e errNoExist) Error() string {
	return fmt.Sprintf("Domain %s not found in gcloud account", e.domain)
}

func (g *gcloud) getZone(domain string) (*gdns.ManagedZone, error) {
	if g.zones == nil {
		g.zones = map[string]*gdns.ManagedZone{}
		pageToken := ""
		for {
			resp, err := g.client.ManagedZones.List(g.project).PageToken(pageToken).Do()
			if err != nil {
				return nil, err
			}
			for _, z := range resp.ManagedZones {
				g.zones[z.DnsName] = z
			}
			if pageToken = resp.NextPageToken; pageToken == "" {
				break
			}
		}
	}
	if g.zones[domain+"."] == nil {
		return nil, errNoExist{domain}
	}
	return g.zones[domain+"."], nil
}

func (g *gcloud) GetNameservers(domain string) ([]*models.Nameserver, error) {
	zone, err := g.getZone(domain)
	if err != nil {
		return nil, err
	}
	return models.StringsToNameservers(zone.NameServers), nil
}

type key struct {
	Type string
	Name string
}

func keyFor(r *gdns.ResourceRecordSet) key {
	return key{Type: r.Type, Name: r.Name}
}
func keyForRec(r *models.RecordConfig) key {
	return key{Type: r.Type, Name: r.NameFQDN + "."}
}

func (g *gcloud) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	if err := dc.Punycode(); err != nil {
		return nil, err
	}
	rrs, zoneName, err := g.getRecords(dc.Name)
	if err != nil {
		return nil, err
	}
	// convert to dnscontrol RecordConfig format
	existingRecords := []*models.RecordConfig{}
	oldRRs := map[key]*gdns.ResourceRecordSet{}
	for _, set := range rrs {
		oldRRs[keyFor(set)] = set
		for _, rec := range set.Rrdatas {
			existingRecords = append(existingRecords, nativeToRecord(set, rec, dc.Name))
		}
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
		panic(errors.Wrap(err, "unparsable record received from GCLOUD"))
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
	fmt.Printf("Adding zone for %s to gcloud account\n", domain)
	mz := &gdns.ManagedZone{
		DnsName:     domain + ".",
		Name:        strings.Replace(domain, ".", "-", -1),
		Description: "zone added by dnscontrol",
	}
	g.zones = nil // reset cache
	_, err = g.client.ManagedZones.Create(g.project, mz).Do()
	return err
}
