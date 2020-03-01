package digitalocean

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/StackExchange/dnscontrol/v2/models"
	"github.com/StackExchange/dnscontrol/v2/providers"
	"github.com/StackExchange/dnscontrol/v2/providers/diff"
	"github.com/miekg/dns/dnsutil"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

/*

DigitalOcean API DNS provider:

Info required in `creds.json`:
   - token

*/

// DoApi is the handle for operations.
type DoApi struct {
	client *godo.Client
}

var defaultNameServerNames = []string{
	"ns1.digitalocean.com",
	"ns2.digitalocean.com",
	"ns3.digitalocean.com",
}

// NewDo creates a DO-specific DNS provider.
func NewDo(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	if m["token"] == "" {
		return nil, fmt.Errorf("no DigitalOcean token provided")
	}

	ctx := context.Background()
	oauthClient := oauth2.NewClient(
		ctx,
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: m["token"]}),
	)
	client := godo.NewClient(oauthClient)

	api := &DoApi{client: client}

	// Get a domain to validate the token
	_, resp, err := api.client.Domains.List(ctx, &godo.ListOptions{PerPage: 1})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token for digitalocean is not valid")
	}

	return api, nil
}

var features = providers.DocumentationNotes{
	providers.DocCreateDomains:       providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	// Digitalocean support CAA records, except
	// ";" value with issue/issuewild records:
	// https://www.digitalocean.com/docs/networking/dns/how-to/create-caa-records/
	providers.CanUseCAA:   providers.Can(),
	providers.CanGetZones: providers.Can(),
}

func init() {
	providers.RegisterDomainServiceProviderType("DIGITALOCEAN", NewDo, features)
}

// EnsureDomainExists returns an error if domain doesn't exist.
func (api *DoApi) EnsureDomainExists(domain string) error {
	ctx := context.Background()
	_, resp, err := api.client.Domains.Get(ctx, domain)
	if resp.StatusCode == http.StatusNotFound {
		_, _, err := api.client.Domains.Create(ctx, &godo.DomainCreateRequest{
			Name:      domain,
			IPAddress: "",
		})
		return err
	}
	return err
}

// GetNameservers returns the nameservers for domain.
func (api *DoApi) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.ToNameservers(defaultNameServerNames)
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *DoApi) GetZoneRecords(domain string) (models.Records, error) {
	records, err := getRecords(api, domain)
	if err != nil {
		return nil, err
	}

	var existingRecords []*models.RecordConfig
	for i := range records {
		r := toRc(domain, &records[i])
		if r.Type == "SOA" {
			continue
		}
		existingRecords = append(existingRecords, r)
	}

	return existingRecords, nil
}

// GetDomainCorrections returns a list of corretions for the  domain.
func (api *DoApi) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	ctx := context.Background()
	dc.Punycode()

	existingRecords, err := api.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	// Normalize
	models.PostProcessRecords(existingRecords)

	differ := diff.New(dc)
	_, create, delete, modify := differ.IncrementalDiff(existingRecords)

	var corrections = []*models.Correction{}

	// Deletes first so changing type works etc.
	for _, m := range delete {
		id := m.Existing.Original.(*godo.DomainRecord).ID
		corr := &models.Correction{
			Msg: fmt.Sprintf("%s, DO ID: %d", m.String(), id),
			F: func() error {
				_, err := api.client.Domains.DeleteRecord(ctx, dc.Name, id)
				return err
			},
		}
		corrections = append(corrections, corr)
	}
	for _, m := range create {
		req := toReq(dc, m.Desired)
		corr := &models.Correction{
			Msg: m.String(),
			F: func() error {
				_, _, err := api.client.Domains.CreateRecord(ctx, dc.Name, req)
				return err
			},
		}
		corrections = append(corrections, corr)
	}
	for _, m := range modify {
		id := m.Existing.Original.(*godo.DomainRecord).ID
		req := toReq(dc, m.Desired)
		corr := &models.Correction{
			Msg: fmt.Sprintf("%s, DO ID: %d", m.String(), id),
			F: func() error {
				_, _, err := api.client.Domains.EditRecord(ctx, dc.Name, id, req)
				return err
			},
		}
		corrections = append(corrections, corr)
	}

	return corrections, nil
}

func getRecords(api *DoApi, name string) ([]godo.DomainRecord, error) {
	ctx := context.Background()

	records := []godo.DomainRecord{}
	opt := &godo.ListOptions{}
	for {
		result, resp, err := api.client.Domains.Records(ctx, name, opt)
		if err != nil {
			return nil, err
		}

		for _, d := range result {
			records = append(records, d)
		}

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return records, nil
}

func toRc(domain string, r *godo.DomainRecord) *models.RecordConfig {
	// This handles "@" etc.
	name := dnsutil.AddOrigin(r.Name, domain)

	target := r.Data
	// Make target FQDN (#rtype_variations)
	if r.Type == "CNAME" || r.Type == "MX" || r.Type == "NS" || r.Type == "SRV" {
		// If target is the domainname, e.g. cname foo.example.com -> example.com,
		// DO returns "@" on read even if fqdn was written.
		if target == "@" {
			target = domain
		} else if target == "." {
			target = ""
		}
		target = target + "."
	}

	t := &models.RecordConfig{
		Type:         r.Type,
		TTL:          uint32(r.TTL),
		MxPreference: uint16(r.Priority),
		SrvPriority:  uint16(r.Priority),
		SrvWeight:    uint16(r.Weight),
		SrvPort:      uint16(r.Port),
		Original:     r,
		CaaTag:       r.Tag,
		CaaFlag:      uint8(r.Flags),
	}
	t.SetLabelFromFQDN(name, domain)
	t.SetTarget(target)
	switch rtype := r.Type; rtype {
	case "TXT":
		t.SetTargetTXTString(target)
	default:
		// nothing additional required
	}
	return t
}

func toReq(dc *models.DomainConfig, rc *models.RecordConfig) *godo.DomainRecordEditRequest {
	name := rc.GetLabel()         // DO wants the short name or "@" for apex.
	target := rc.GetTargetField() // DO uses the target field only for a single value
	priority := 0                 // DO uses the same property for MX and SRV priority

	switch rc.Type { // #rtype_variations
	case "MX":
		priority = int(rc.MxPreference)
	case "SRV":
		priority = int(rc.SrvPriority)
	case "TXT":
		// TXT records are the one place where DO combines many items into one field.
		target = rc.GetTargetCombined()
	case "CAA":
		// DO API requires that value ends in dot
		// But the value returned from API doesn't contain this,
		// so no need to strip the dot when reading value from API.
		target = target + "."
	default:
		// no action required
	}

	return &godo.DomainRecordEditRequest{
		Type:     rc.Type,
		Name:     name,
		Data:     target,
		TTL:      int(rc.TTL),
		Priority: priority,
		Port:     int(rc.SrvPort),
		Weight:   int(rc.SrvWeight),
		Tag:      rc.CaaTag,
		Flags:    int(rc.CaaFlag),
	}
}
