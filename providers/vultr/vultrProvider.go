package vultr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/vultr/govultr"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

/*

Vultr API DNS provider:

Info required in `creds.json`:
   - token

*/

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
}

func init() {
	providers.RegisterDomainServiceProviderType("VULTR", NewProvider, features)
}

// vultrProvider represents the Vultr DNSServiceProvider.
type vultrProvider struct {
	client *govultr.Client
	token  string
}

// defaultNS contains the default nameservers for Vultr.
var defaultNS = []string{
	"ns1.vultr.com",
	"ns2.vultr.com",
}

// NewProvider initializes a Vultr DNSServiceProvider.
func NewProvider(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	token := m["token"]
	if token == "" {
		return nil, fmt.Errorf("missing Vultr API token")
	}

	client := govultr.NewClient(nil, token)
	client.SetUserAgent("dnscontrol")

	_, err := client.Account.GetInfo(context.Background())
	return &vultrProvider{client, token}, err
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *vultrProvider) GetZoneRecords(domain string) (models.Records, error) {
	records, err := api.client.DNSRecord.List(context.Background(), domain)
	if err != nil {
		return nil, err
	}

	curRecords := make(models.Records, len(records))
	for i := range records {
		r, err := toRecordConfig(domain, &records[i])
		if err != nil {
			return nil, err
		}
		curRecords[i] = r
	}

	return curRecords, nil
}

// GetDomainCorrections gets the corrections for a DomainConfig.
func (api *vultrProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()

	curRecords, err := api.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	models.PostProcessRecords(curRecords)

	differ := diff.New(dc)
	_, create, delete, modify, err := differ.IncrementalDiff(curRecords)
	if err != nil {
		return nil, err
	}

	var corrections []*models.Correction

	for _, mod := range delete {
		id := mod.Existing.Original.(*govultr.DNSRecord).RecordID
		corrections = append(corrections, &models.Correction{
			Msg: fmt.Sprintf("%s; Vultr RecordID: %v", mod.String(), id),
			F: func() error {
				return api.client.DNSRecord.Delete(context.Background(), dc.Name, strconv.Itoa(id))
			},
		})
	}

	for _, mod := range create {
		r := toVultrRecord(dc, mod.Desired, 0)
		corrections = append(corrections, &models.Correction{
			Msg: mod.String(),
			F: func() error {
				return api.client.DNSRecord.Create(context.Background(), dc.Name, r.Type, r.Name, r.Data, r.TTL, vultrPriority(r))
			},
		})
	}

	for _, mod := range modify {
		r := toVultrRecord(dc, mod.Desired, mod.Existing.Original.(*govultr.DNSRecord).RecordID)
		corrections = append(corrections, &models.Correction{
			Msg: fmt.Sprintf("%s; Vultr RecordID: %v", mod.String(), r.RecordID),
			F: func() error {
				return api.client.DNSRecord.Update(context.Background(), dc.Name, r)
			},
		})
	}

	return corrections, nil
}

// GetNameservers gets the Vultr nameservers for a domain
func (api *vultrProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.ToNameservers(defaultNS)
}

// EnsureDomainExists adds a domain to the Vutr DNS service if it does not exist
func (api *vultrProvider) EnsureDomainExists(domain string) error {
	if ok, err := api.isDomainInAccount(domain); err != nil {
		return err
	} else if ok {
		return nil
	}

	// Vultr requires an initial IP, use a dummy one.
	return api.client.DNSDomain.Create(context.Background(), domain, "0.0.0.0")
}

func (api *vultrProvider) isDomainInAccount(domain string) (bool, error) {
	domains, err := api.client.DNSDomain.List(context.Background())
	if err != nil {
		return false, err
	}
	for _, d := range domains {
		if d.Domain == domain {
			return true, nil
		}
	}
	return false, nil
}

// toRecordConfig converts a Vultr DNSRecord to a RecordConfig. #rtype_variations
func toRecordConfig(domain string, r *govultr.DNSRecord) (*models.RecordConfig, error) {
	origin, data := domain, r.Data
	rc := &models.RecordConfig{
		TTL:      uint32(r.TTL),
		Original: r,
	}
	rc.SetLabel(r.Name, domain)

	switch rtype := r.Type; rtype {
	case "CNAME", "NS":
		rc.Type = r.Type
		// Make target into a FQDN if it is a CNAME, NS, MX, or SRV.
		if !strings.HasSuffix(data, ".") {
			data = data + "."
		}
		return rc, rc.SetTarget(data)
	case "CAA":
		// Vultr returns CAA records in the format "[flag] [tag] [value]".
		return rc, rc.SetTargetCAAString(data)
	case "MX":
		if !strings.HasSuffix(data, ".") {
			data = data + "."
		}
		return rc, rc.SetTargetMX(uint16(vultrPriority(r)), data)
	case "SRV":
		// Vultr returns SRV records in the format "[weight] [port] [target]".
		return rc, rc.SetTargetSRVPriorityString(uint16(vultrPriority(r)), data)
	case "TXT":
		// Remove quotes if it is a TXT record.
		if !strings.HasPrefix(data, `"`) || !strings.HasSuffix(data, `"`) {
			return nil, errors.New("unexpected lack of quotes in TXT record from Vultr")
		}
		return rc, rc.SetTargetTXT(data[1 : len(data)-1])
	default:
		return rc, rc.PopulateFromString(rtype, r.Data, origin)
	}
}

// toVultrRecord converts a RecordConfig converted by toRecordConfig back to a Vultr DNSRecord. #rtype_variations
func toVultrRecord(dc *models.DomainConfig, rc *models.RecordConfig, vultrID int) *govultr.DNSRecord {
	name := rc.GetLabel()
	// Vultr uses a blank string to represent the apex domain.
	if name == "@" {
		name = ""
	}

	data := rc.GetTargetField()

	// Vultr does not use a period suffix for CNAME, NS, or MX.
	if strings.HasSuffix(data, ".") {
		data = data[:len(data)-1]
	}
	// Vultr needs TXT record in quotes.
	if rc.Type == "TXT" {
		data = fmt.Sprintf(`"%s"`, data)
	}

	var priority *int

	if rc.Type == "MX" {
		tmp := int(rc.MxPreference)
		priority = &tmp
	}
	if rc.Type == "SRV" {
		tmp := int(rc.SrvPriority)
		priority = &tmp
	}

	r := &govultr.DNSRecord{
		RecordID: vultrID,
		Type:     rc.Type,
		Name:     name,
		Data:     data,
		TTL:      int(rc.TTL),
		Priority: priority,
	}
	switch rtype := rc.Type; rtype { // #rtype_variations
	case "SRV":
		r.Data = fmt.Sprintf("%v %v %s", rc.SrvWeight, rc.SrvPort, rc.GetTargetField())
	case "CAA":
		r.Data = fmt.Sprintf(`%v %s "%s"`, rc.CaaFlag, rc.CaaTag, rc.GetTargetField())
	case "SSHFP":
		r.Data = fmt.Sprintf("%d %d %s", rc.SshfpAlgorithm, rc.SshfpFingerprint, rc.GetTargetField())
	default:
	}

	return r
}

func vultrPriority(r *govultr.DNSRecord) int {
	if r.Priority == nil {
		return 0
	}
	return *r.Priority
}
