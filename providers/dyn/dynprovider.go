package dyn

import (
	"encoding/json"
	"fmt"
	"log"

	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/miekg/dns/dnsutil"
	"github.com/nesv/go-dynect/dynect"
)

type dynProvider struct {
	*dynect.ConvenientClient
}

func init() {
	providers.RegisterDomainServiceProviderType("DYN", create)
}

func create(creds map[string]string, meta json.RawMessage) (providers.DNSServiceProvider, error) {
	d := &dynProvider{}
	customer := creds["customer"]
	user := creds["username"]
	pass := creds["password"]
	if customer == "" || user == "" || pass == "" {
		return nil, fmt.Errorf("DYN requires customer, username, and password")
	}
	cli := dynect.NewClient(customer)
	d.ConvenientClient = &dynect.ConvenientClient{Client: *cli}
	err := d.Login(user, pass)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (d *dynProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.StringsToNameservers([]string{"ns1.p04.dynect.net", "ns2.p04.dynect.net", "ns3.p04.dynect.net", "ns4.p04.dynect.net"}), nil
}

func (d *dynProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	// Preprocess records to accomidate Provider-specific quirks
	dc.Punycode()
	dc.Filter(func(r *models.RecordConfig) bool {
		// DYN does not let you modify apex NS record TTLs or content
		if r.Type == "NS" && r.Name == "@" {
			r.TTL = 3600 * 24
		}
		return true
	})

	recs, err := d.getRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	df := diff.New(dc)
	_, creates, deletes, mods := df.IncrementalDiff(recs)

	corrections := []*models.Correction{}
	for _, dl := range deletes {
		corrections = append(corrections, &models.Correction{
			Msg: dl.String(),
			F:   d.del(dl.Existing.Original.(*dynect.BaseRecord), dc.Name),
		})
	}
	for _, c := range creates {
		corrections = append(corrections, &models.Correction{
			Msg: c.String(),
			F:   d.create(c.Desired, dc.Name),
		})
	}
	for _, m := range mods {
		corrections = append(corrections, &models.Correction{
			Msg: m.String(),
			F:   d.modify(m.Existing.Original.(*dynect.BaseRecord).RecordId, m.Desired, dc.Name),
		})
	}
	if len(corrections) > 0 {
		corrections = append(corrections, &models.Correction{
			Msg: "Publish zone",
			F:   func() error { return d.PublishZone(dc.Name) },
		})
	}
	return corrections, nil
}

func (d *dynProvider) create(r *models.RecordConfig, domain string) func() error {
	return func() error {
		resource := fmt.Sprintf("%sRecord/%s/%s", r.Type, domain, r.NameFQDN)
		rec := buildRecord(r, domain)
		resp := &dynect.ResponseBlock{}
		err := d.Do("POST", resource, rec, resp)
		return err
	}
}

func (d *dynProvider) modify(id int, r *models.RecordConfig, domain string) func() error {
	return func() error {
		resource := fmt.Sprintf("%sRecord/%s/%s/%d", r.Type, domain, r.NameFQDN, id)
		rec := buildRecord(r, domain)
		resp := &dynect.ResponseBlock{}
		err := d.Do("PUT", resource, rec, resp)
		return err
	}
}

func (d *dynProvider) del(r *dynect.BaseRecord, domain string) func() error {
	return func() error {
		resource := fmt.Sprintf("%sRecord/%s/%s/%d", r.RecordType, domain, r.FQDN, r.RecordId)
		resp := &dynect.ResponseBlock{}
		err := d.Do("DELETE", resource, nil, resp)
		return err
	}
}

func (d *dynProvider) getRecords(domain string) ([]*models.RecordConfig, error) {
	recs := []*models.RecordConfig{}
	resp := &dynect.AllRecordsResponse{}
	err := d.Do("GET", "AllRecord/"+domain, nil, resp)
	if err != nil {
		return nil, err
	}
	for _, ref := range resp.Data {
		if !strings.HasPrefix(ref, "/REST/") {
			return nil, fmt.Errorf("Unexpected record reference detected: %s", ref)
		}
		// trim "/REST/"
		ref = ref[6:]

		rec := &dynect.RecordResponse{}
		err = d.Do("GET", ref, nil, rec)
		if err != nil {
			return nil, err
		}
		r := convertToRecord(&rec.Data, domain)
		if r != nil {
			recs = append(recs, r)
		}
	}
	return recs, nil
}

func buildRecord(r *models.RecordConfig, domain string) *dynect.RecordRequest {
	rec := &dynect.RecordRequest{
		TTL: fmt.Sprint(r.TTL),
	}
	switch r.Type {
	case "A", "AAAA":
		rec.RData.Address = r.Target
	case "NS":
		rec.RData.NSDName = r.Target
	case "CNAME":
		rec.RData.CName = r.Target
	case "MX":
		rec.RData.Exchange = r.Target
		rec.RData.Preference = int(r.Priority)
	case "TXT":
		rec.RData.TxtData = r.Target
	default:
		panic("OOOOOOO")
	}
	return rec
}

func convertToRecord(r *dynect.BaseRecord, domain string) *models.RecordConfig {
	content := ""
	var priority uint16
	switch r.RecordType {
	case "A":
		content = r.RData.Address
	case "NS":
		content = r.RData.NSDName
	case "CNAME":
		content = r.RData.CName
	case "MX":
		content = r.RData.Exchange
		priority = uint16(r.RData.Preference)
	case "TXT":
		content = r.RData.TxtData
	case "SOA":
		return nil
	default:
		log.Printf("DYN provider does not know how to process record type %s for %s.", r.RecordType, r.FQDN)
		return nil
	}
	return &models.RecordConfig{
		NameFQDN: r.FQDN,
		Name:     dnsutil.TrimDomainName(r.FQDN, domain),
		Original: r,
		Type:     r.RecordType,
		TTL:      uint32(r.TTL),
		Priority: priority,
		Target:   content,
	}
}
