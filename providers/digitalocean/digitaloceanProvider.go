package digitalocean

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/miekg/dns/dnsutil"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

/*

Digitalocean API DNS provider:

Info required in `creds.json`:
   - token

*/

type DoApi struct {
	client *godo.Client
}

var defaultNameServerNames = []string{
	"ns1.digitalocean.com",
	"ns2.digitalocean.com",
	"ns3.digitalocean.com",
}

func NewDo(m map[string]string, metadata json.RawMessage) (models.DNSServiceProviderDriver, error) {
	if m["token"] == "" {
		return nil, fmt.Errorf("Digitalocean Token must be provided.")
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
		return nil, fmt.Errorf("Digitalocean Token is not valid.")
	}

	return api, nil
}

var docNotes = providers.DocumentationNotes{
	providers.DocCreateDomains:       providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	providers.RegisterDomainServiceProviderType("DIGITALOCEAN", NewDo, providers.CanUseSRV, docNotes)
}

func (api *DoApi) EnsureDomainExists(domain string) error {
	ctx := context.Background()
	_, resp, err := api.client.Domains.Get(ctx, domain)
	if resp.StatusCode == http.StatusNotFound {
		_, _, err := api.client.Domains.Create(ctx, &godo.DomainCreateRequest{
			Name:      domain,
			IPAddress: "",
		})
		return err
	} else {
		return err
	}
}

func (api *DoApi) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.StringsToNameservers(defaultNameServerNames), nil
}

func (api *DoApi) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	ctx := context.Background()
	dc.Punycode()

	records, err := getRecords(api, dc.Name)
	if err != nil {
		return nil, err
	}

	existingRecords := make([]*models.RecordConfig, len(records))
	for i := range records {
		existingRecords[i] = toRc(dc, &records[i])
	}

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

func toRc(dc *models.DomainConfig, r *godo.DomainRecord) *models.RecordConfig {
	// This handles "@" etc.
	name := dnsutil.AddOrigin(r.Name, dc.Name)

	target := r.Data
	// Make target FQDN (#rtype_variations)
	if r.Type == "CNAME" || r.Type == "MX" || r.Type == "NS" || r.Type == "SRV" {
		// If target is the domainname, e.g. cname foo.example.com -> example.com,
		// DO returns "@" on read even if fqdn was written.
		if target == "@" {
			target = dc.Name
		}
		target = dnsutil.AddOrigin(target+".", dc.Name)
	}

	return &models.RecordConfig{
		NameFQDN:     name,
		Type:         r.Type,
		Target:       target,
		TTL:          uint32(r.TTL),
		MxPreference: uint16(r.Priority),
		SrvPriority:  uint16(r.Priority),
		SrvWeight:    uint16(r.Weight),
		SrvPort:      uint16(r.Port),
		Original:     r,
	}
}

func toReq(dc *models.DomainConfig, rc *models.RecordConfig) *godo.DomainRecordEditRequest {
	// DO wants the short name, e.g. @
	name := dnsutil.TrimDomainName(rc.NameFQDN, dc.Name)

	// DO uses the same property for MX and SRV priority
	priority := 0
	switch rc.Type { // #rtype_variations
	case "MX":
		priority = int(rc.MxPreference)
	case "SRV":
		priority = int(rc.SrvPriority)
	}

	return &godo.DomainRecordEditRequest{
		Type:     rc.Type,
		Name:     name,
		Data:     rc.Target,
		TTL:      int(rc.TTL),
		Priority: priority,
		Port:     int(rc.SrvPort),
		Weight:   int(rc.SrvWeight),
	}
}
