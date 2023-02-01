package netlify

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/miekg/dns"
)

var nameServerSuffixes = []string{
	".nsone.net.",
}

var features = providers.DocumentationNotes{
	providers.CanAutoDNSSEC:          providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Cannot(),
	providers.CanUseDSForChildren:    providers.Cannot(),
	providers.CanUseNAPTR:            providers.Cannot(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Cannot(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot("Netlify does not allow sufficient control over the apex NS records"),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   newNetlify,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("NETLIFY", fns, features)
	providers.RegisterCustomRecordType("NETLIFY", "NETLIFY", "")
	providers.RegisterCustomRecordType("NETLIFYv6", "NETLIFY", "")
}

type netlifyProvider struct {
	apiToken    string // the account access token
	accountSlug string // the account identifier slug. optional.
}

func newNetlify(m map[string]string, message json.RawMessage) (providers.DNSServiceProvider, error) {
	api := &netlifyProvider{}
	api.apiToken = m["token"]
	if api.apiToken == "" {
		return nil, fmt.Errorf("missing Netlify personal access token")
	}

	api.accountSlug = m["slug"]

	return api, nil
}

func (n *netlifyProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	zone, err := n.getZone(domain)
	if err != nil {
		return nil, err
	}

	return models.ToNameservers(zone.DNSServers)
}

func (n *netlifyProvider) getZone(domain string) (*dnsZone, error) {
	zones, err := n.getDNSZones()
	if err != nil {
		return nil, err
	}

	for _, zone := range zones {
		if zone.Name == domain {
			return zone, nil
		}
	}

	return nil, fmt.Errorf("no zones found for this domain")
}

func (n *netlifyProvider) GetZoneRecords(domain string) (models.Records, error) {
	zone, err := n.getZone(domain)
	if err != nil {
		return nil, err
	}

	records, err := n.getDNSRecords(zone.ID)
	if err != nil {
		return nil, err
	}

	cleanRecords := make(models.Records, 0)

	for _, r := range records {
		if r.Type == "SOA" {
			continue
		}

		rec := &models.RecordConfig{
			TTL:      uint32(r.TTL),
			Original: r,
		}

		rec.SetLabelFromFQDN(r.Hostname, domain) // netlify returns the FQDN

		if r.Type == "CNAME" || r.Type == "MX" || r.Type == "NS" {
			r.Value = dns.CanonicalName(r.Value)
		}

		switch rtype := r.Type; rtype {
		case "NETLIFY", "NETLIFYv6": // transparently ignore
			continue
		case "MX":
			err = rec.SetTargetMX(uint16(r.Priority), r.Value)
		case "SRV":
			parts := strings.Fields(r.Value)
			if len(parts) == 3 {
				r.Value += "."
			}
			err = rec.SetTargetSRV(uint16(r.Priority), r.Weight, r.Port, r.Value)
		case "TXT":
			err = rec.SetTargetTXT(r.Value)
		case "CAA":
			err = rec.SetTargetCAA(uint8(r.Flag), r.Tag, r.Value)
		default:
			err = rec.PopulateFromString(r.Type, r.Value, domain)
		}

		if err != nil {
			return nil, fmt.Errorf("unparsable record received from Netlify: %w", err)
		}

		cleanRecords = append(cleanRecords, rec)
	}

	return cleanRecords, nil
}

// Return true if the string ends in one of Netlify's name server domains
// False if anything else
func isNetlifyNameServerDomain(name string) bool {
	for _, i := range nameServerSuffixes {
		if strings.HasSuffix(name, i) {
			return true
		}
	}
	return false
}

// remove all non-netlify NS records from our desired state.
// if any are found, print a warning
func removeOtherApexNS(dc *models.DomainConfig) {
	newList := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.Type == "NS" {
			// apex NS inside netlify are expected.
			// We ignore them, warning as needed.
			// Child delegations are supported so, we allow non-apex NS records.
			if rec.GetLabelFQDN() == dc.Name {
				if !isNetlifyNameServerDomain(rec.GetTargetField()) {
					printer.Printf("Warning: Netlify does not allow NS records to be modified. %s will not be added.\n", rec.GetTargetField())
				}
				continue
			}
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}

func (n *netlifyProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {

	err := dc.Punycode()
	if err != nil {
		return nil, err
	}

	records, err := n.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	// Normalize
	models.PostProcessRecords(records)
	removeOtherApexNS(dc)

	var corrections []*models.Correction
	var create, del, modify diff.Changeset
	if !diff2.EnableDiff2 {
		differ := diff.New(dc)
		_, create, del, modify, err = differ.IncrementalDiff(records)
	} else {
		differ := diff.NewCompat(dc)
		_, create, del, modify, err = differ.IncrementalDiff(records)
	}
	if err != nil {
		return nil, err
	}

	zone, err := n.getZone(dc.Name)
	if err != nil {
		return nil, err
	}

	// Deletes first so changing type works etc.
	for _, m := range del {
		id := m.Existing.Original.(*dnsRecord).ID
		corr := &models.Correction{
			Msg: m.String(),
			F: func() error {
				return n.deleteDNSRecord(zone.ID, id)
			},
		}
		corrections = append(corrections, corr)
	}

	for _, m := range create {
		req := toReq(m.Desired)
		corr := &models.Correction{
			Msg: m.String(),
			F: func() error {
				_, err := n.createDNSRecord(zone.ID, req)
				return err
			},
		}
		corrections = append(corrections, corr)
	}

	for _, m := range modify {
		id := m.Existing.Original.(*dnsRecord).ID
		req := toReq(m.Desired)
		corr := &models.Correction{
			Msg: m.String(),
			F: func() error {
				if err := n.deleteDNSRecord(zone.ID, id); err != nil {
					return err
				}

				_, err := n.createDNSRecord(zone.ID, req)
				return err
			},
		}
		corrections = append(corrections, corr)
	}

	return corrections, nil
}

func toReq(rc *models.RecordConfig) *dnsRecordCreate {
	name := rc.GetLabelFQDN() // Netlify wants the FQDN
	target := rc.GetTargetField()
	priority := int64(0)

	switch rc.Type {
	case "MX":
		priority = int64(rc.MxPreference)
	case "SRV":
		priority = int64(rc.SrvPriority)
	case "TXT":
		target = rc.GetTargetTXTJoined()
	default:
		// no action required
	}

	return &dnsRecordCreate{
		Type:     rc.Type,
		Hostname: name,
		Value:    target,
		TTL:      int64(rc.TTL),
		Priority: priority,
		Port:     int64(rc.SrvPort),
		Weight:   int64(rc.SrvWeight),
		Tag:      rc.CaaTag,
		Flag:     int64(rc.CaaFlag),
	}
}
