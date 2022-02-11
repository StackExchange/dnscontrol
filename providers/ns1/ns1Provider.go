package ns1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
	"gopkg.in/ns1/ns1-go.v2/rest/model/filter"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

var docNotes = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.CanGetZones:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   newProvider,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("NS1", fns, providers.CanUseSRV, docNotes)
	providers.RegisterCustomRecordType("NS1_URLFWD", "NS1", "URLFWD")
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

func (n *nsone) EnsureDomainExists(domain string) error {
	// This enables the create-domains subcommand

	zone := dns.NewZone(domain)
	_, err := n.Zones.Create(zone)

	if err == rest.ErrZoneExists {
		// if domain exists already, just return nil, nothing to do here.
		return nil
	}

	return err
}

func (n *nsone) GetNameservers(domain string) ([]*models.Nameserver, error) {
	z, _, err := n.Zones.Get(domain)
	if err != nil {
		return nil, err
	}
	return models.ToNameservers(z.DNSServers)
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (n *nsone) GetZoneRecords(domain string) (models.Records, error) {
	z, _, err := n.Zones.Get(domain)
	if err != nil {
		return nil, err
	}

	found := models.Records{}
	for _, r := range z.Records {
		zrs, err := convert(r, domain)
		if err != nil {
			return nil, err
		}
		found = append(found, zrs...)
	}
	return found, nil
}

func (n *nsone) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()
	//dc.CombineMXs()

	domain := dc.Name

	// Get existing records
	existingRecords, err := n.GetZoneRecords(domain)
	if err != nil {
		return nil, err
	}

	existingGrouped := existingRecords.GroupedByKey()
	desiredGrouped := dc.Records.GroupedByKey()

	//  Normalize
	models.PostProcessRecords(existingRecords)

	differ := diff.New(dc)
	changedGroups, err := differ.ChangedGroups(existingRecords)
	if err != nil {
		return nil, err
	}
	corrections := []*models.Correction{}
	// each name/type is given to the api as a unit.
	for k, descs := range changedGroups {
		key := k

		desc := strings.Join(descs, "\n")
		_, current := existingGrouped[k]
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
	_, err := n.Records.Delete(domain, key.NameFQDN, key.Type)
	return err
}

func (n *nsone) modify(recs models.Records, domain string) error {
	_, err := n.Records.Update(buildRecord(recs, domain, ""))
	return err
}

func buildRecord(recs models.Records, domain string, id string) *dns.Record {
	r := recs[0]
	rec := &dns.Record{
		Domain:  r.GetLabelFQDN(),
		Type:    r.Type,
		ID:      id,
		TTL:     int(r.TTL),
		Zone:    domain,
		Filters: []*filter.Filter{}, // Work through a bug in the NS1 API library that causes 400 Input validation failed (Value None for field '<obj>.filters' is not of type array)
	}
	for _, r := range recs {
		if r.Type == "MX" {
			rec.AddAnswer(&dns.Answer{Rdata: strings.Split(fmt.Sprintf("%d %v", r.MxPreference, r.GetTargetField()), " ")})
		} else if r.Type == "TXT" {
			rec.AddAnswer(&dns.Answer{Rdata: r.TxtStrings})
		} else if r.Type == "CAA" {
			rec.AddAnswer(&dns.Answer{
				Rdata: []string{
					fmt.Sprintf("%v", r.CaaFlag),
					r.CaaTag,
					fmt.Sprintf("%s", r.GetTargetField()),
			}})
		} else if r.Type == "SRV" {
			rec.AddAnswer(&dns.Answer{Rdata: strings.Split(fmt.Sprintf("%d %d %d %v", r.SrvPriority, r.SrvWeight, r.SrvPort, r.GetTargetField()), " ")})
		} else if r.Type == "NAPTR" {
			rec.AddAnswer(&dns.Answer{Rdata: []string{
				fmt.Sprintf("%d", r.NaptrOrder),
				fmt.Sprintf("%d", r.NaptrPreference),
				r.NaptrFlags,
				r.NaptrService,
				r.NaptrRegexp,
				r.GetTargetField()}})
		} else if r.Type == "TLSA" {
			rec.AddAnswer(&dns.Answer{Rdata: []string{
				fmt.Sprintf("%d", r.TlsaUsage),
				fmt.Sprintf("%d", r.TlsaSelector),
				fmt.Sprintf("%d", r.TlsaMatchingType),
				r.GetTargetField()}})
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
		case "ALIAS":
			rec.Type = rtype
			if err := rec.SetTarget(ans); err != nil {
				return nil, fmt.Errorf("unparsable %s record received from ns1: %w", rtype, err)
			}
		case "URLFWD":
			rec.Type = rtype
			if err := rec.SetTarget(ans); err != nil {
				return nil, fmt.Errorf("unparsable %s record received from ns1: %w", rtype, err)
			}
		case "CAA":
			//dnscontrol expects quotes around multivalue CAA entries, API doesn't add them
			x_ans := strings.SplitN(ans, " ", 3)
			if err := rec.SetTargetCAAStrings(x_ans[0], x_ans[1], x_ans[2]); err != nil {
				return nil, fmt.Errorf("unparsable %s record received from ns1: %w", rtype, err)
			}
		default:
			if err := rec.PopulateFromString(rtype, ans, domain); err != nil {
				return nil, fmt.Errorf("unparsable record received from ns1: %w", err)
			}
		}
		found = append(found, rec)
	}
	return found, nil
}
