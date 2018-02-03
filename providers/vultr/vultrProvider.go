package vultr

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/miekg/dns/dnsutil"

	vultr "github.com/JamesClonk/vultr/lib"
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
	providers.DocCreateDomains:       providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	providers.RegisterDomainServiceProviderType("VULTR", NewVultr, features)
}

// VultrApi represents the Vultr DNSServiceProvider
type VultrApi struct {
	client *vultr.Client
	token  string
}

// defaultNS are the default nameservers for Vultr
var defaultNS = []string{
	"ns1.vultr.com",
	"ns2.vultr.com",
}

// NewVultr initializes a Vultr DNSServiceProvider
func NewVultr(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	api := &VultrApi{
		token: m["token"],
	}

	if api.token == "" {
		return nil, fmt.Errorf("Vultr API token is required")
	}

	api.client = vultr.NewClient(api.token, nil)

	// Validate token
	_, err := api.client.GetAccountInfo()
	if err != nil {
		return nil, err
	}

	return api, nil
}

// GetDomainCorrections gets the corrections for a DomainConfig
func (api *VultrApi) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()

	ok, err := api.isDomainInAccount(dc.Name)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("%s is not a domain in the Vultr account", dc.Name)
	}

	records, err := api.client.GetDNSRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	curRecords := make([]*models.RecordConfig, len(records))
	for i := range records {
		r, err := toRecordConfig(dc, &records[i])
		if err != nil {
			return nil, err
		}

		curRecords[i] = r
	}

	// Normalize
	models.PostProcessRecords(curRecords)

	differ := diff.New(dc)
	_, create, delete, modify := differ.IncrementalDiff(curRecords)

	corrections := []*models.Correction{}

	for _, mod := range delete {
		id := mod.Existing.Original.(*vultr.DNSRecord).RecordID
		corrections = append(corrections, &models.Correction{
			Msg: fmt.Sprintf("%s; Vultr RecordID: %v", mod.String(), id),
			F: func() error {
				return api.client.DeleteDNSRecord(dc.Name, id)
			},
		})
	}

	for _, mod := range create {
		r := toVultrRecord(dc, mod.Desired)
		corrections = append(corrections, &models.Correction{
			Msg: mod.String(),
			F: func() error {
				return api.client.CreateDNSRecord(dc.Name, r.Name, r.Type, r.Data, r.Priority, r.TTL)
			},
		})
	}

	for _, mod := range modify {
		id := mod.Existing.Original.(*vultr.DNSRecord).RecordID
		r := toVultrRecord(dc, mod.Desired)
		r.RecordID = id
		corrections = append(corrections, &models.Correction{
			Msg: fmt.Sprintf("%s; Vultr RecordID: %v", mod.String(), id),
			F: func() error {
				return api.client.UpdateDNSRecord(dc.Name, *r)
			},
		})
	}

	return corrections, nil
}

// GetNameservers gets the Vultr nameservers for a domain
func (api *VultrApi) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.StringsToNameservers(defaultNS), nil
}

// EnsureDomainExists adds a domain to the Vutr DNS service if it does not exist
func (api *VultrApi) EnsureDomainExists(domain string) error {
	ok, err := api.isDomainInAccount(domain)
	if err != nil {
		return err
	}

	if !ok {
		// Vultr requires an initial IP, use a dummy one
		err := api.client.CreateDNSDomain(domain, "127.0.0.1")
		if err != nil {
			return err
		}

		ok, err := api.isDomainInAccount(domain)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("Unexpected error adding domain %s to Vultr account", domain)
		}
	}

	return nil
}

func (api *VultrApi) isDomainInAccount(domain string) (bool, error) {
	domains, err := api.client.GetDNSDomains()
	if err != nil {
		return false, err
	}

	var vd *vultr.DNSDomain
	for _, d := range domains {
		if d.Domain == domain {
			vd = &d
		}
	}

	if vd == nil {
		return false, nil
	}

	return true, nil
}

// toRecordConfig converts a Vultr DNSRecord to a RecordConfig #rtype_variations
func toRecordConfig(dc *models.DomainConfig, r *vultr.DNSRecord) (*models.RecordConfig, error) {
	// Turns r.Name into a FQDN
	// Vultr uses "" as the apex domain, instead of "@", and this handles it fine.
	name := dnsutil.AddOrigin(r.Name, dc.Name)

	data := r.Data
	// Make target into a FQDN if it is a CNAME, NS, MX, or SRV
	if r.Type == "CNAME" || r.Type == "NS" || r.Type == "MX" {
		if !strings.HasSuffix(data, ".") {
			data = data + "."
		}
		data = dnsutil.AddOrigin(data, dc.Name)
	}
	// Remove quotes if it is a TXT
	if r.Type == "TXT" {
		if !strings.HasPrefix(data, `"`) || !strings.HasSuffix(data, `"`) {
			return nil, errors.New("Unexpected lack of quotes in TXT record from Vultr")
		}
		data = data[1 : len(data)-1]
	}

	rc := &models.RecordConfig{
		NameFQDN: name,
		Type:     r.Type,
		Target:   data,
		TTL:      uint32(r.TTL),
		Original: r,
	}

	if r.Type == "MX" {
		rc.MxPreference = uint16(r.Priority)
	}

	if r.Type == "SRV" {
		rc.SrvPriority = uint16(r.Priority)

		// Vultr returns in the format "[weight] [port] [target]"
		splitData := strings.SplitN(rc.Target, " ", 3)
		if len(splitData) != 3 {
			return nil, fmt.Errorf("Unexpected data for SRV record returned by Vultr")
		}

		weight, err := strconv.ParseUint(splitData[0], 10, 16)
		if err != nil {
			return nil, err
		}
		rc.SrvWeight = uint16(weight)

		port, err := strconv.ParseUint(splitData[1], 10, 16)
		if err != nil {
			return nil, err
		}
		rc.SrvPort = uint16(port)

		target := splitData[2]
		if !strings.HasSuffix(target, ".") {
			target = target + "."
		}
		rc.Target = dnsutil.AddOrigin(target, dc.Name)
	}

	if r.Type == "CAA" {
		// Vultr returns in the format "[flag] [tag] [value]"

		splitData := strings.SplitN(rc.Target, " ", 3)
		if len(splitData) != 3 {
			return nil, fmt.Errorf("Unexpected data for CAA record returned by Vultr")
		}

		flag, err := strconv.ParseUint(splitData[0], 10, 8)
		if err != nil {
			return nil, err
		}
		rc.CaaFlag = uint8(flag)

		rc.CaaTag = splitData[1]

		value := splitData[2]
		if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
			value = value[1 : len(value)-1]
		}
		if strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`) {
			value = value[1 : len(value)-1]
		}
		rc.Target = value
	}

	return rc, nil
}

// toVultrRecord converts a RecordConfig converted by toRecordConfig back to a Vultr DNSRecord #rtype_variations
func toVultrRecord(dc *models.DomainConfig, rc *models.RecordConfig) *vultr.DNSRecord {
	name := dnsutil.TrimDomainName(rc.NameFQDN, dc.Name)

	// Vultr uses a blank string to represent the apex domain
	if name == "@" {
		name = ""
	}

	data := rc.Target

	// Vultr does not use a period suffix for the server for CNAME, NS, or MX
	if strings.HasSuffix(data, ".") {
		data = data[:len(data)-1]
	}
	// Vultr needs TXT record in quotes
	if rc.Type == "TXT" {
		data = fmt.Sprintf(`"%s"`, data)
	}

	priority := 0

	if rc.Type == "MX" {
		priority = int(rc.MxPreference)
	}

	if rc.Type == "SRV" {
		priority = int(rc.SrvPriority)
	}

	r := &vultr.DNSRecord{
		Type:     rc.Type,
		Name:     name,
		Data:     data,
		TTL:      int(rc.TTL),
		Priority: priority,
	}

	if rc.Type == "SRV" {
		target := rc.Target
		if strings.HasSuffix(target, ".") {
			target = target[:len(target)-1]
		}

		r.Data = fmt.Sprintf("%v %v %s", rc.SrvWeight, rc.SrvPort, target)
	}

	if rc.Type == "CAA" {
		r.Data = fmt.Sprintf(`%v %s "%s"`, rc.CaaFlag, rc.CaaTag, rc.Target)
	}

	return r
}
