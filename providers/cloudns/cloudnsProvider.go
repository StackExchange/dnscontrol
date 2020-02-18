package cloudns

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/miekg/dns/dnsutil"

	"github.com/StackExchange/dnscontrol/v2/models"
	"github.com/StackExchange/dnscontrol/v2/providers"
	"github.com/StackExchange/dnscontrol/v2/providers/diff"
)

/*

CloDNS API DNS provider:

Info required in `creds.json`:
   - auth-id
   - auth-password

*/

// NewCloudns creates the provider.
func NewCloudns(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	c := &api{}

	c.creds.id, c.creds.password = m["auth-id"], m["auth-password"]
	if c.creds.id == "" || c.creds.password == "" {
		return nil, fmt.Errorf("missing ClouDNS auth-id and auth-password")
	}

	// Get a domain to validate authentication
	if err := c.fetchDomainList(); err != nil {
		return nil, err
	}

	return c, nil
}

var features = providers.DocumentationNotes{
	providers.DocDualHost:            providers.Unimplemented(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.DocCreateDomains:       providers.Can(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.CanUsePTR:              providers.Unimplemented(),
	providers.CanGetZones:            providers.Unimplemented(),
}

func init() {
	providers.RegisterDomainServiceProviderType("CLOUDNS", NewCloudns, features)
}

// GetNameservers returns the nameservers for a domain.
func (c *api) GetNameservers(domain string) ([]*models.Nameserver, error) {
	if len(c.nameserversNames) == 0 {
		c.fetchAvailableNameservers()
	}
	return models.StringsToNameservers(c.nameserversNames), nil
}

// GetDomainCorrections returns the corrections for a domain.
func (c *api) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc, err := dc.Copy()
	if err != nil {
		return nil, err
	}

	dc.Punycode()

	if c.domainIndex == nil {
		if err := c.fetchDomainList(); err != nil {
			return nil, err
		}
	}
	domainID, ok := c.domainIndex[dc.Name]
	if !ok {
		return nil, fmt.Errorf("'%s' not a zone in ClouDNS account", dc.Name)
	}

	records, err := c.getRecords(domainID)
	if err != nil {
		return nil, err
	}

	existingRecords := make([]*models.RecordConfig, len(records), len(records)+len(c.nameserversNames))
	for i := range records {
		existingRecords[i] = toRc(dc, &records[i])
	}
	// Normalize
	models.PostProcessRecords(existingRecords)

	// ClouDNS doesn't allow selecting an arbitrary TTL, only a set of predefined values https://asia.cloudns.net/wiki/article/188/
	// We need to make sure we don't change it every time if it is as close as it's going to get
	for _, record := range dc.Records {
		record.TTL = fixTTL(record.TTL)
	}

	differ := diff.New(dc)
	_, create, del, modify := differ.IncrementalDiff(existingRecords)

	var corrections []*models.Correction

	// Deletes first so changing type works etc.
	for _, m := range del {
		id := m.Existing.Original.(*domainRecord).ID
		corr := &models.Correction{
			Msg: fmt.Sprintf("%s, ClouDNS ID: %s", m.String(), id),
			F: func() error {
				return c.deleteRecord(domainID, id)
			},
		}
		corrections = append(corrections, corr)
	}

	for _, m := range create {
		req, err := toReq(m.Desired)
		if err != nil {
			return nil, err
		}

		corr := &models.Correction{
			Msg: m.String(),
			F: func() error {
				return c.createRecord(domainID, req)
			},
		}
		corrections = append(corrections, corr)
	}
	for _, m := range modify {
		id := m.Existing.Original.(*domainRecord).ID
		req, err := toReq(m.Desired)
		if err != nil {
			return nil, err
		}

		corr := &models.Correction{
			Msg: fmt.Sprintf("%s, ClouDNS ID: %s: ", m.String(), id),
			F: func() error {
				return c.modifyRecord(domainID, id, req)
			},
		}
		corrections = append(corrections, corr)
	}

	return corrections, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (client *api) GetZoneRecords(domain string) (models.Records, error) {
	return nil, fmt.Errorf("not implemented")
	// This enables the get-zones subcommand.
	// Implement this by extracting the code from GetDomainCorrections into
	// a single function.  For most providers this should be relatively easy.
}

// EnsureDomainExists returns an error if domain doesn't exist.
func (c *api) EnsureDomainExists(domain string) error {
	if err := c.fetchDomainList(); err != nil {
		return err
	}
	// domain already exists
	if _, ok := c.domainIndex[domain]; ok {
		return nil
	}
	return c.createDomain(domain)
}

func toRc(dc *models.DomainConfig, r *domainRecord) *models.RecordConfig {

	ttl, _ := strconv.ParseUint(r.TTL, 10, 32)
	priority, _ := strconv.ParseUint(r.Priority, 10, 32)
	weight, _ := strconv.ParseUint(r.Weight, 10, 32)
	port, _ := strconv.ParseUint(r.Port, 10, 32)

	rc := &models.RecordConfig{
		Type:         r.Type,
		TTL:          uint32(ttl),
		MxPreference: uint16(priority),
		SrvPriority:  uint16(priority),
		SrvWeight:    uint16(weight),
		SrvPort:      uint16(port),
		Original:     r,
	}
	rc.SetLabel(r.Host, dc.Name)

	switch rtype := r.Type; rtype { // #rtype_variations
	case "TXT":
		rc.SetTargetTXT(r.Target)
	case "CNAME", "MX", "NS", "SRV", "ALIAS":
		rc.SetTarget(dnsutil.AddOrigin(r.Target+".", dc.Name))
	case "CAA":
		caaFlag, _ := strconv.ParseUint(r.CaaFlag, 10, 32)
		rc.CaaFlag = uint8(caaFlag)
		rc.CaaTag = r.CaaTag
		rc.SetTarget(r.CaaValue)
	case "TLSA":
		tlsaUsage, _ := strconv.ParseUint(r.TlsaUsage, 10, 32)
		rc.TlsaUsage = uint8(tlsaUsage)
		tlsaSelector, _ := strconv.ParseUint(r.TlsaSelector, 10, 32)
		rc.TlsaSelector = uint8(tlsaSelector)
		tlsaMatchingType, _ := strconv.ParseUint(r.TlsaMatchingType, 10, 32)
		rc.TlsaMatchingType = uint8(tlsaMatchingType)
		rc.SetTarget(r.Target)
	case "SSHFP":
		sshfpAlgorithm, _ := strconv.ParseUint(r.SshfpAlgorithm, 10, 32)
		rc.SshfpAlgorithm = uint8(sshfpAlgorithm)
		sshfpFingerprint, _ := strconv.ParseUint(r.SshfpFingerprint, 10, 32)
		rc.SshfpFingerprint = uint8(sshfpFingerprint)
		rc.SetTarget(r.Target)
	default:
		rc.SetTarget(r.Target)
	}

	return rc
}

func toReq(rc *models.RecordConfig) (requestParams, error) {
	req := requestParams{
		"record-type": rc.Type,
		"host":        rc.GetLabel(),
		"record":      rc.GetTargetField(),
		"ttl":         strconv.Itoa(int(rc.TTL)),
	}

	// ClouDNS doesn't use "@", it uses an empty name
	if req["host"] == "@" {
		req["host"] = ""
	}

	switch rc.Type { // #rtype_variations
	case "A", "AAAA", "NS", "PTR", "TXT", "SOA", "ALIAS", "CNAME":
		// Nothing special.
	case "MX":
		req["priority"] = strconv.Itoa(int(rc.MxPreference))
	case "SRV":
		req["priority"] = strconv.Itoa(int(rc.SrvPriority))
		req["weight"] = strconv.Itoa(int(rc.SrvWeight))
		req["port"] = strconv.Itoa(int(rc.SrvPort))
	case "CAA":
		req["caa_flag"] = strconv.Itoa(int(rc.CaaFlag))
		req["caa_type"] = rc.CaaTag
		req["caa_value"] = rc.Target
	case "TLSA":
		req["tlsa_usage"] = strconv.Itoa(int(rc.TlsaUsage))
		req["tlsa_selector"] = strconv.Itoa(int(rc.TlsaSelector))
		req["tlsa_matching_type"] = strconv.Itoa(int(rc.TlsaMatchingType))
	case "SSHFP":
		req["algorithm"] = strconv.Itoa(int(rc.SshfpAlgorithm))
		req["fptype"] = strconv.Itoa(int(rc.SshfpFingerprint))
	default:
		msg := fmt.Sprintf("ClouDNS.toReq rtype %v unimplemented", rc.Type)
		panic(msg)
		// We panic so that we quickly find any switch statements
	}

	return req, nil
}
