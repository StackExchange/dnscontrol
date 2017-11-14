package linode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/miekg/dns/dnsutil"

	"net/url"

	"golang.org/x/oauth2"
	"regexp"
	"strings"
)

/*

Linode API DNS provider:

Info required in `creds.json`:
   - token

*/

var allowedTTLValues = []uint32{
	300,     // 5 minutes
	3600,    // 1 hour
	7200,    // 2 hours
	14400,   // 4 hours
	28800,   // 8 hours
	57600,   // 16 hours
	86400,   // 1 day
	172800,  // 2 days
	345600,  // 4 days
	604800,  // 1 week
	1209600, // 2 weeks
	2419200, // 4 weeks
}

var srvRegexp = regexp.MustCompile(`^_(?P<Service>\w+)\.\_(?P<Protocol>\w+)$`)

type LinodeApi struct {
	client      *http.Client
	baseURL     *url.URL
	domainIndex map[string]int
}

var defaultNameServerNames = []string{
	"ns1.linode.com",
	"ns2.linode.com",
	"ns3.linode.com",
	"ns4.linode.com",
	"ns5.linode.com",
}

func NewLinode(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	if m["token"] == "" {
		return nil, fmt.Errorf("Linode Token must be provided.")
	}

	ctx := context.Background()
	client := oauth2.NewClient(
		ctx,
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: m["token"]}),
	)

	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, fmt.Errorf("Linode base URL not valid")
	}

	api := &LinodeApi{client: client, baseURL: baseURL}

	// Get a domain to validate the token
	if err := api.fetchDomainList(); err != nil {
		return nil, err
	}

	return api, nil
}

var docNotes = providers.DocumentationNotes{
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.DocDualHost:            providers.Cannot(),
}

func init() {
	// SRV support is in this provider, but Linode doesn't seem to support it properly
	providers.RegisterDomainServiceProviderType("LINODE", NewLinode, docNotes)
}

func (api *LinodeApi) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.StringsToNameservers(defaultNameServerNames), nil
}

func (api *LinodeApi) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc, err := dc.Copy()
	if err != nil {
		return nil, err
	}

	dc.Punycode()

	if api.domainIndex == nil {
		if err := api.fetchDomainList(); err != nil {
			return nil, err
		}
	}
	domainID, ok := api.domainIndex[dc.Name]
	if !ok {
		return nil, fmt.Errorf("%s not listed in domains for Linode account", dc.Name)
	}

	records, err := api.getRecords(domainID)
	if err != nil {
		return nil, err
	}

	existingRecords := make([]*models.RecordConfig, len(records), len(records)+len(defaultNameServerNames))
	for i := range records {
		existingRecords[i] = toRc(dc, &records[i])
	}

	// Linode always has read-only NS servers, but these are not mentioned in the API response
	// https://github.com/linode/manager/blob/edd99dc4e1be5ab8190f243c3dbf8b830716255e/src/constants.js#L184
	for _, name := range defaultNameServerNames {
		existingRecords = append(existingRecords, &models.RecordConfig{
			NameFQDN: dc.Name,
			Type:     "NS",
			Target:   name,
			Original: &domainRecord{},
		})
	}

	// Normalize
	models.Downcase(existingRecords)

	// Linode doesn't allow selecting an arbitrary TTL, only a set of predefined values
	// We need to make sure we don't change it every time if it is as close as it's going to get
	// By experimentation, Linode always rounds up. 300 -> 300, 301 -> 3600.
	// https://github.com/linode/manager/blob/edd99dc4e1be5ab8190f243c3dbf8b830716255e/src/domains/components/SelectDNSSeconds.js#L19
	for _, record := range dc.Records {
		record.TTL = fixTTL(record.TTL)
	}

	differ := diff.New(dc)
	_, create, del, modify := differ.IncrementalDiff(existingRecords)

	var corrections []*models.Correction

	// Deletes first so changing type works etc.
	for _, m := range del {
		id := m.Existing.Original.(*domainRecord).ID
		if id == 0 { // Skip ID 0, these are the default nameservers always present
			continue
		}
		corr := &models.Correction{
			Msg: fmt.Sprintf("%s, Linode ID: %d", m.String(), id),
			F: func() error {
				return api.deleteRecord(domainID, id)
			},
		}
		corrections = append(corrections, corr)
	}
	for _, m := range create {
		req, err := toReq(dc, m.Desired)
		if err != nil {
			return nil, err
		}
		j, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		corr := &models.Correction{
			Msg: fmt.Sprintf("%s: %s", m.String(), string(j)),
			F: func() error {
				record, err := api.createRecord(domainID, req)
				if err != nil {
					return err
				}
				// TTL isn't saved when creating a record, so we will need to modify it immediately afterwards
				return api.modifyRecord(domainID, record.ID, req)
			},
		}
		corrections = append(corrections, corr)
	}
	for _, m := range modify {
		id := m.Existing.Original.(*domainRecord).ID
		if id == 0 { // Skip ID 0, these are the default nameservers always present
			continue
		}
		req, err := toReq(dc, m.Desired)
		if err != nil {
			return nil, err
		}
		j, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		corr := &models.Correction{
			Msg: fmt.Sprintf("%s, Linode ID: %d: %s", m.String(), id, string(j)),
			F: func() error {
				return api.modifyRecord(domainID, id, req)
			},
		}
		corrections = append(corrections, corr)
	}

	return corrections, nil
}

func toRc(dc *models.DomainConfig, r *domainRecord) *models.RecordConfig {
	// This handles "@" etc.
	name := dnsutil.AddOrigin(r.Name, dc.Name)

	target := r.Target
	// Make target FQDN (#rtype_variations)
	if r.Type == "CNAME" || r.Type == "MX" || r.Type == "NS" || r.Type == "SRV" {
		target = dnsutil.AddOrigin(target+".", dc.Name)
	}

	return &models.RecordConfig{
		NameFQDN:     name,
		Type:         r.Type,
		Target:       target,
		TTL:          r.TTLSec,
		MxPreference: r.Priority,
		SrvPriority:  r.Priority,
		SrvWeight:    r.Weight,
		SrvPort:      uint16(r.Port),
		Original:     r,
	}
}

func toReq(dc *models.DomainConfig, rc *models.RecordConfig) (*recordEditRequest, error) {
	req := &recordEditRequest{
		Type:     rc.Type,
		Name:     dnsutil.TrimDomainName(rc.NameFQDN, dc.Name),
		Target:   rc.Target,
		TTL:      int(rc.TTL),
		Priority: 0,
		Port:     int(rc.SrvPort),
		Weight:   int(rc.SrvWeight),
	}

	// Linode doesn't use "@", it uses an empty name
	if req.Name == "@" {
		req.Name = ""
	}

	// Linode uses the same property for MX and SRV priority
	switch rc.Type { // #rtype_variations
	case "A", "AAAA", "NS", "PTR", "TXT", "SOA", "TLSA", "CAA":
		// Nothing special.
	case "MX":
		req.Priority = int(rc.MxPreference)
		req.Target = fixTarget(req.Target, dc.Name)
	case "SRV":
		req.Priority = int(rc.SrvPriority)

		// From softlayer provider
		// This is to support SRV, it doesn't work yet for Linode
		result := srvRegexp.FindStringSubmatch(req.Name)

		if len(result) != 3 {
			return nil, fmt.Errorf("SRV Record must match format \"_service._protocol\" not %s", req.Name)
		}

		var serviceName, protocol string = result[1], strings.ToLower(result[2])

		req.Protocol = protocol
		req.Service = serviceName
		req.Name = ""
	case "CNAME":
		req.Target = fixTarget(req.Target, dc.Name)
	default:
		msg := fmt.Sprintf("linode.toReq rtype %v unimplemented", rc.Type)
		panic(msg)
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}

	return req, nil
}

func fixTarget(target, domain string) string {
	// Linode always wants a fully qualified target name
	if target[len(target)-1] == '.' {
		return target[:len(target)-1]
	} else {
		return fmt.Sprintf("%s.%s", target, domain)
	}
}

func fixTTL(ttl uint32) uint32 {
	// if the TTL is larger than the largest allowed value, return the largest allowed value
	if ttl > allowedTTLValues[len(allowedTTLValues)-1] {
		return allowedTTLValues[len(allowedTTLValues)-1]
	}

	for _, v := range allowedTTLValues {
		if v >= ttl {
			return v
		}
	}

	return allowedTTLValues[0]
}
