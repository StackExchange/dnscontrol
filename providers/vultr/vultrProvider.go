package vultr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/oauth2"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/vultr/govultr/v2"
)

/*

Vultr API DNS provider:

Info required in `creds.json`:
   - token

*/

var features = providers.DocumentationNotes{
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   NewProvider,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("VULTR", fns, features)
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

	config := &oauth2.Config{}

	client := govultr.NewClient(config.Client(context.Background(), &oauth2.Token{AccessToken: token}))
	client.SetUserAgent("dnscontrol")

	_, err := client.Account.Get(context.Background())
	return &vultrProvider{client, token}, err
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *vultrProvider) GetZoneRecords(domain string) (models.Records, error) {
	listOptions := &govultr.ListOptions{}
	records, meta, err := api.client.DomainRecord.List(context.Background(), domain, listOptions)
	curRecords := make(models.Records, meta.Total)
	nextI := 0

	for {
		if err != nil {
			return nil, err
		}
		currentI := 0
		for i, record := range records {
			r, err := toRecordConfig(domain, record)
			if err != nil {
				return nil, err
			}
			curRecords[nextI+i] = r
			currentI = nextI + i
		}
		nextI = currentI + 1

		if meta.Links.Next == "" {
			break
		} else {
			listOptions.Cursor = meta.Links.Next
			records, meta, err = api.client.DomainRecord.List(context.Background(), domain, listOptions)
			continue
		}
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
		id := mod.Existing.Original.(govultr.DomainRecord).ID
		corrections = append(corrections, &models.Correction{
			Msg: fmt.Sprintf("%s; Vultr RecordID: %v", mod.String(), id),
			F: func() error {
				return api.client.DomainRecord.Delete(context.Background(), dc.Name, id)
			},
		})
	}

	for _, mod := range create {
		r := toVultrRecord(dc, mod.Desired, "0")
		corrections = append(corrections, &models.Correction{
			Msg: mod.String(),
			F: func() error {
				_, err := api.client.DomainRecord.Create(context.Background(), dc.Name, &govultr.DomainRecordReq{Name: r.Name, Type: r.Type, Data: r.Data, TTL: r.TTL, Priority: &r.Priority})
				return err
			},
		})
	}

	for _, mod := range modify {
		r := toVultrRecord(dc, mod.Desired, mod.Existing.Original.(govultr.DomainRecord).ID)
		corrections = append(corrections, &models.Correction{
			Msg: fmt.Sprintf("%s; Vultr RecordID: %v", mod.String(), r.ID),
			F: func() error {
				return api.client.DomainRecord.Update(context.Background(), dc.Name, r.ID, &govultr.DomainRecordReq{Name: r.Name, Type: r.Type, Data: r.Data, TTL: r.TTL, Priority: &r.Priority})
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
	_, err := api.client.Domain.Create(context.Background(), &govultr.DomainReq{Domain: domain, IP: "0.0.0.0", DNSSec: "disabled"})
	return err
}

func (api *vultrProvider) isDomainInAccount(domain string) (bool, error) {
	listOptions := &govultr.ListOptions{}
	domains, meta, err := api.client.Domain.List(context.Background(), listOptions)

	for {
		if err != nil {
			return false, err
		}

		for _, d := range domains {
			if d.Domain == domain {
				return true, nil
			}
		}

		if meta.Links.Next == "" {
			break
		} else {
			listOptions.Cursor = meta.Links.Next
			domains, meta, err = api.client.Domain.List(context.Background(), listOptions)
			continue
		}
	}
	return false, nil
}

// toRecordConfig converts a Vultr DomainRecord to a RecordConfig. #rtype_variations
func toRecordConfig(domain string, r govultr.DomainRecord) (*models.RecordConfig, error) {
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
		return rc, rc.SetTargetMX(uint16(r.Priority), data)
	case "SRV":
		// Vultr returns SRV records in the format "[weight] [port] [target]".
		return rc, rc.SetTargetSRVPriorityString(uint16(r.Priority), data)
	case "TXT":
		// TXT records from Vultr are always surrounded by quotes.
		// They don't permit quotes within the string, therefore there is no
		// need to resolve \" or other quoting.
		if !(strings.HasPrefix(data, `"`) && strings.HasSuffix(data, `"`)) {
			// Give an error if Vultr changes their protocol. We'd rather break
			// than do the wrong thing.
			return nil, errors.New("unexpected lack of quotes in TXT record from Vultr")
		}
		return rc, rc.SetTargetTXT(data[1 : len(data)-1])
	default:
		return rc, rc.PopulateFromString(rtype, r.Data, origin)
	}
}

// toVultrRecord converts a RecordConfig converted by toRecordConfig back to a Vultr DomainRecordReq. #rtype_variations
func toVultrRecord(dc *models.DomainConfig, rc *models.RecordConfig, vultrID string) *govultr.DomainRecord {
	name := rc.GetLabel()
	// Vultr uses a blank string to represent the apex domain.
	if name == "@" {
		name = ""
	}

	data := rc.GetTargetField()

	// Vultr does not use a period suffix for CNAME, NS, or MX.
	data = strings.TrimSuffix(data, ".")

	priority := 0

	if rc.Type == "MX" {
		priority = int(rc.MxPreference)
	}
	if rc.Type == "SRV" {
		priority = int(rc.SrvPriority)
	}

	r := &govultr.DomainRecord{
		ID:       vultrID,
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
	case "TXT":
		// Vultr doesn't permit TXT strings to include double-quotes
		// therefore, we don't have to escape interior double-quotes.
		// Vultr's API requires the string to begin and end with double-quotes.
		r.Data = `"` + strings.Join(rc.TxtStrings, "") + `"`
	default:
	}

	return r
}
