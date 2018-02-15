package ns1

import (
	"encoding/json"

	"fmt"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/pkg/errors"

	"net/http"

	"strings"

	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/miekg/dns/dnsutil"
	"gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

var docNotes = providers.DocumentationNotes{
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.DocDualHost:            providers.Can(),
}

func init() {
	providers.RegisterDomainServiceProviderType("NS1", newProvider, providers.CanUseSRV, docNotes)
}

type nsone struct {
	*rest.Client
}

func newProvider(creds map[string]string, meta json.RawMessage) (providers.DNSServiceProvider, error) {
	if creds["api_token"] == "" {
		return nil, errors.Errorf("api_token required for ns1")
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
	//dc.CombineMXs()
	z, _, err := n.Zones.Get(dc.Name)
	if err != nil {
		return nil, err
	}

	found := models.Records{}
	for _, r := range z.Records {
		zrs, err := convert(r, dc.Name)
		if err != nil {
			return nil, err
		}
		found = append(found, zrs...)
	}
	foundGrouped := found.Grouped()
	desiredGrouped := dc.Records.Grouped()

	//  Normalize
	models.PostProcessRecords(found)

	differ := diff.New(dc)
	changedGroups := differ.ChangedGroups(found)
	corrections := []*models.Correction{}
	// each name/type is given to the api as a unit.
	for k, descs := range changedGroups {
		key := k
		desc := strings.Join(descs, "\n")
		_, current := foundGrouped[k]
		recs, wanted := desiredGrouped[k]
		if wanted && !current {
			// pure addition
			corrections = append(corrections, &models.Correction{
				Msg: desc,
				F:   func() error { return n.add(recs, dc.Name) },
			})
		} else if current && !wanted {
			// pure deletion
			corrections = append(corrections, &models.Correction{
				Msg: desc,
				F:   func() error { return n.remove(key, dc.Name) },
			})
		} else {
			// modification
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
		Domain: r.GetLabelFQDN(),
		Type:   r.Type,
		ID:     id,
		TTL:    int(r.TTL),
		Zone:   domain,
	}
	for _, r := range recs {
		if r.Type == "TXT" {
			rec.AddAnswer(&dns.Answer{Rdata: r.TxtStrings})
		} else if r.Type == "SRV" {
			rec.AddAnswer(&dns.Answer{Rdata: strings.Split(fmt.Sprintf("%d %d %d %v", r.SrvPriority, r.SrvWeight, r.SrvPort, r.GetTargetField()), " ")})
		} else {
			rec.AddAnswer(&dns.Answer{Rdata: strings.Split(r.GetTargetField(), " ")})
		}
	}
	return rec
}

func convert(zr *dns.ZoneRecord, domain string) ([]*models.RecordConfig, error) {
	found := []*models.RecordConfig{}
	for _, ans := range zr.ShortAns {
		rec := &models.RecordConfig{
			TTL:      uint32(zr.TTL),
			Original: zr,
		}
		rec.SetLabelFromFQDN(zr.Domain, domain)
		switch rtype := zr.Type; rtype {
		default:
			if err := rec.PopulateFromString(rtype, ans, domain); err != nil {
				panic(errors.Wrap(err, "unparsable record received from ns1"))
			}
		}
		found = append(found, rec)
	}
	return found, nil
}
