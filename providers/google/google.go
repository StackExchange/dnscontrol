package google

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"

	"github.com/StackExchange/dnscontrol/providers/diff"
	"golang.org/x/oauth2"
	gauth "golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"
	"strings"
)

func init() {
	providers.RegisterDomainServiceProviderType("GCLOUD", New)
}

type gcloud struct {
	client  *dns.Service
	project string
	zones   map[string]*dns.ManagedZone
}

func New(cfg map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	for _, key := range []string{"clientId", "clientSecret", "refreshToken", "project"} {
		if cfg[key] == "" {
			return nil, fmt.Errorf("%s required for google cloud provider", key)
		}
	}
	ocfg := &oauth2.Config{
		Endpoint:     gauth.Endpoint,
		ClientID:     cfg["clientId"],
		ClientSecret: cfg["clientSecret"],
	}
	tok := &oauth2.Token{
		RefreshToken: cfg["refreshToken"],
	}
	client := ocfg.Client(context.Background(), tok)
	dcli, err := dns.New(client)
	if err != nil {
		return nil, err
	}
	return &gcloud{
		client:  dcli,
		project: cfg["project"],
	}, nil
}

func (g *gcloud) getZone(domain string) (*dns.ManagedZone, error) {
	if g.zones == nil {
		g.zones = map[string]*dns.ManagedZone{}
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
		return nil, fmt.Errorf("Domain %s not found in gcloud account", domain)
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

func keyFor(r *dns.ResourceRecordSet) key {
	return key{Type: r.Type, Name: r.Name}
}
func keyForRec(r *models.RecordConfig) key {
	return key{Type: r.Type, Name: r.NameFQDN + "."}
}

func (g *gcloud) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	rrs, zoneName, err := g.getRecords(dc.Name)
	if err != nil {
		return nil, err
	}
	//convert to dnscontrol RecordConfig format
	existingRecords := []diff.Record{}
	oldRRs := map[key]*dns.ResourceRecordSet{}
	for _, set := range rrs {
		nameWithoutDot := set.Name
		if strings.HasSuffix(nameWithoutDot, ".") {
			nameWithoutDot = nameWithoutDot[:len(nameWithoutDot)-1]
		}
		oldRRs[keyFor(set)] = set
		for _, rec := range set.Rrdatas {
			r := &models.RecordConfig{
				NameFQDN: nameWithoutDot,
				Type:     set.Type,
				Target:   rec,
				TTL:      uint32(set.Ttl),
			}
			existingRecords = append(existingRecords, r)
		}
	}

	w := []diff.Record{}
	for _, want := range dc.Records {
		if want.TTL == 0 {
			want.TTL = 300
		}
		if want.Type == "MX" {
			want.Target = fmt.Sprintf("%d %s", want.Priority, want.Target)
			want.Priority = 0
		} else if want.Type == "TXT" {
			//add quotes to txts
			want.Target = fmt.Sprintf(`"%s"`, want.Target)
		}
		w = append(w, want)
	}

	// first collect keys that have changed
	_, create, delete, modify := diff.IncrementalDiff(existingRecords, w)
	changedKeys := map[key]bool{}
	desc := ""
	for _, c := range create {
		desc += fmt.Sprintln(c)
		changedKeys[keyForRec(c.Desired.(*models.RecordConfig))] = true
	}
	for _, d := range delete {
		desc += fmt.Sprintln(d)
		changedKeys[keyForRec(d.Existing.(*models.RecordConfig))] = true
	}
	for _, m := range modify {
		desc += fmt.Sprintln(m)
		changedKeys[keyForRec(m.Existing.(*models.RecordConfig))] = true
	}
	if len(changedKeys) == 0 {
		return nil, nil
	}
	chg := &dns.Change{Kind: "dns#change"}
	for ck := range changedKeys {
		// remove old version (if present)
		if old, ok := oldRRs[ck]; ok {
			chg.Deletions = append(chg.Deletions, old)
		}
		//collect records to replace with
		newRRs := &dns.ResourceRecordSet{
			Name: ck.Name,
			Type: ck.Type,
			Kind: "dns#resourceRecordSet",
		}
		for _, r := range dc.Records {
			if keyForRec(r) == ck {
				newRRs.Rrdatas = append(newRRs.Rrdatas, r.Target)
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
	return []*models.Correction{
		{
			Msg: desc,
			F:   runChange,
		},
	}, nil
}

func (g *gcloud) getRecords(domain string) ([]*dns.ResourceRecordSet, string, error) {
	zone, err := g.getZone(domain)
	if err != nil {
		return nil, "", err
	}
	pageToken := ""
	sets := []*dns.ResourceRecordSet{}
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
