package ns1

import (
	"encoding/json"

	"fmt"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"

	"net/http"

	"strings"

	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/miekg/dns/dnsutil"
	"gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

func init() {
	providers.RegisterDomainServiceProviderType("NS1", newProvider)
}

type nsone struct {
	*rest.Client
}

func newProvider(creds map[string]string, meta json.RawMessage) (providers.DNSServiceProvider, error) {
	if creds["api_token"] == "" {
		return nil, fmt.Errorf("api_token required for ns1")
	}
	return &nsone{rest.NewClient(http.DefaultClient, rest.SetAPIKey(creds["api_token"]))}, nil
}

func (n *nsone) GetNameservers(domain string) ([]*models.Nameserver, error) {
	z, _, err := n.Zones.Get(domain)
	if err != nil {
		return nil, err
	}
	return models.StringsToNameservers(z.DNSServers), nil
}

func (n *nsone) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()
	dc.CombineMXs()
	z, _, err := n.Zones.Get(dc.Name)
	if err != nil {
		return nil, err
	}

	found := models.Records{}
	for _, r := range z.Records {
		found = append(found, convert(r, dc.Name)...)
	}
	foundGrouped := found.Grouped()
	desiredGrouped := dc.Records.Grouped()

	differ := diff.New(dc)
	changedGroups := differ.ChangedGroups(found)
	corrections := []*models.Correction{}
	for k, descs := range changedGroups {
		desc := strings.Join(descs, "\n")
		_, current := foundGrouped[k]
		recs, wanted := desiredGrouped[k]
		if wanted && !current {
			corrections = append(corrections, &models.Correction{
				Msg: desc,
				F:   func() error { return n.add(recs, dc.Name) },
			})
		} else if current && !wanted {
			corrections = append(corrections, &models.Correction{
				Msg: desc,
				F:   func() error { return n.remove(k, dc.Name) },
			})
		} else {
			corrections = append(corrections, &models.Correction{
				Msg: desc,
				F:   func() error { return n.modify(recs, dc.Name) },
			})
		}
	}

	return corrections, nil
}

func (n *nsone) add(recs models.Records, domain string) error {
	_, err := n.Records.Create(buildRecord(recs, domain, ""))
	return err
}

func (n *nsone) remove(key models.RecordKey, domain string) error {
	_, err := n.Records.Delete(domain, dnsutil.AddOrigin(key.Name, domain), key.Type)
	return err
}

func (n *nsone) modify(recs models.Records, domain string) error {
	_, err := n.Records.Update(buildRecord(recs, domain, ""))
	return err
}

func buildRecord(recs models.Records, domain string, id string) *dns.Record {
	r := recs[0]
	rec := &dns.Record{
		Domain: r.NameFQDN,
		Type:   r.Type,
		ID:     id,
		TTL:    int(r.TTL),
		Zone:   domain,
	}

	for _, r := range recs {
		ans := &dns.Answer{
			Rdata: []string{r.Target},
		}
		rec.AddAnswer(ans)
	}
	return rec
}

func convert(zr *dns.ZoneRecord, domain string) []*models.RecordConfig {
	found := []*models.RecordConfig{}
	for _, ans := range zr.ShortAns {
		rec := &models.RecordConfig{
			NameFQDN: zr.Domain,
			Name:     dnsutil.TrimDomainName(zr.Domain, domain),
			TTL:      uint32(zr.TTL),
			Target:   ans,
			Original: zr,
			Type:     zr.Type,
		}
		found = append(found, rec)
	}
	return found
}
