package digitalocean

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/miekg/dns/dnsutil"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

/*

DigitalOcean API DNS provider:

Info required in `creds.json`:
   - token

*/

// digitaloceanProvider is the handle for operations.
type digitaloceanProvider struct {
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

	api := &digitaloceanProvider{client: client}

	// Get a domain to validate the token
retry:
	_, resp, err := api.client.Domains.List(ctx, &godo.ListOptions{PerPage: 1})
	if err != nil {
		if pauseAndRetry(resp) {
			goto retry
		}
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
func (api *digitaloceanProvider) EnsureDomainExists(domain string) error {
retry:
	ctx := context.Background()
	_, resp, err := api.client.Domains.Get(ctx, domain)
	if err != nil {
		if pauseAndRetry(resp) {
			goto retry
		}
		//return err
	}
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
func (api *digitaloceanProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.ToNameservers(defaultNameServerNames)
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *digitaloceanProvider) GetZoneRecords(domain string) (models.Records, error) {
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
func (api *digitaloceanProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	ctx := context.Background()
	dc.Punycode()

	existingRecords, err := api.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	// Normalize
	models.PostProcessRecords(existingRecords)

	differ := diff.New(dc)
	_, create, delete, modify, err := differ.IncrementalDiff(existingRecords)
	if err != nil {
		return nil, err
	}

	var corrections = []*models.Correction{}

	// Deletes first so changing type works etc.
	for _, m := range delete {
		id := m.Existing.Original.(*godo.DomainRecord).ID
		corr := &models.Correction{
			Msg: fmt.Sprintf("%s, DO ID: %d", m.String(), id),
			F: func() error {
			retry:
				resp, err := api.client.Domains.DeleteRecord(ctx, dc.Name, id)
				if err != nil {
					if pauseAndRetry(resp) {
						goto retry
					}
				}
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
			retry:
				_, resp, err := api.client.Domains.CreateRecord(ctx, dc.Name, req)
				if err != nil {
					if pauseAndRetry(resp) {
						goto retry
					}
				}
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
			retry:
				_, resp, err := api.client.Domains.EditRecord(ctx, dc.Name, id, req)
				if err != nil {
					if pauseAndRetry(resp) {
						goto retry
					}
				}
				return err
			},
		}
		corrections = append(corrections, corr)
	}

	return corrections, nil
}

func getRecords(api *digitaloceanProvider, name string) ([]godo.DomainRecord, error) {
	ctx := context.Background()

retry:

	records := []godo.DomainRecord{}
	opt := &godo.ListOptions{}
	for {
		result, resp, err := api.client.Domains.Records(ctx, name, opt)
		if err != nil {
			if pauseAndRetry(resp) {
				goto retry
			}
			return nil, err
		}

		records = append(records, result...)

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

// backoff is the amount of time to sleep if a 429 or 504 is received.
// It is doubled after each use.
var backoff = time.Second * 5

/*
This is what a resp looks like when a retry is required:

&{Status:429 Too Many Requests StatusCode:429 Proto:HTTP/2.0 ProtoMajor:2 ProtoMinor:0 Header:map[Cf-Cache-Status:[DYNAMIC] Cf-Ray:[5ee9cb277bf2c5fc-EWR] Cf-Request-Id:[0645fd4caa0000c5fcdb3ca000000001] Content-Length:[60] Content-Type:[application/json] Date:[Sat, 07 Nov 2020 20:27:11 GMT] Expect-Ct:[max-age=604800, report-uri="https://report-uri.cloudflare.com/cdn-cgi/beacon/expect-ct"] Ratelimit-Limit:[5000] Ratelimit-Remaining:[0] Ratelimit-Reset:[1604779129] Server:[cloudflare] Set-Cookie:[__cfduid=dcf76c52eb750b2e64c8e0e7f738289e11604780831; expires=Mon, 07-Dec-20 20:27:11 GMT; path=/; domain=.digitalocean.com; HttpOnly; SameSite=Lax] X-Gateway:[Edge-Gateway] X-Request-Id:[b331028a-9388-4198-9608-2c931ba66827] X-Response-From:[Edge-Gateway]] Body:{cs:0xc0009c3b80} ContentLength:60 TransferEncoding:[] Close:false Uncompressed:false Trailer:map[] Request:0xc0008fa900 TLS:0xc0004a0160}
*/

func pauseAndRetry(resp *godo.Response) bool {
	statusCode := resp.Response.StatusCode
	if statusCode != 429 && statusCode != 504 {
		return false
	}

	// Did the API return a specific time to try again?
	var rateLimitReset int64
	rr := resp.Header["Ratelimit-Reset"]
	if len(rr) == 1 {
		var err error
		rateLimitReset, err = strconv.ParseInt(rr[0], 10, 64)
		if err != nil {
			rateLimitReset = 0
		}
	}
	if rateLimitReset != 0 {
		f := time.Unix(rateLimitReset, 0).UTC()
		log.Printf("Rate limit exceeded. Ratelimit-Reset is %v %v\n", rateLimitReset, f)
		d := time.Until(f)
		log.Printf("TIME dur: %v\n", d)
		if d > time.Second {
			time.Sleep(d)
			return true
		}
	}

	// Just use a simple exponential back-off with a 2-minute max.
	log.Printf("Using exponential backoff instead: %v seconds\n", backoff)
	time.Sleep(backoff)
	if backoff < (time.Second * 120) {
		backoff = backoff * 2
	}
	return true
}
