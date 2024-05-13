package ns1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
	"gopkg.in/ns1/ns1-go.v2/rest/model/filter"
)

var docNotes = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanAutoDNSSEC:          providers.Can(),
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Can(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDNAME:            providers.Can(),
	providers.CanUseDS:               providers.Can(),
	providers.CanUseDSForChildren:    providers.Can(),
	providers.CanUseDHCID:            providers.Can(),
	providers.CanUseHTTPS:            providers.Can(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSVCB:             providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

// clientRetries is the number of retries for API backend requests in case of StatusTooManyRequests responses
const clientRetries = 10

func init() {
	fns := providers.DspFuncs{
		Initializer:   newProvider,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("NS1", fns, providers.CanUseSRV, docNotes)
	providers.RegisterCustomRecordType("NS1_URLFWD", "NS1", "")
}

type nsone struct {
	*rest.Client
}

func newProvider(creds map[string]string, meta json.RawMessage) (providers.DNSServiceProvider, error) {
	if creds["api_token"] == "" {
		return nil, fmt.Errorf("api_token required for ns1")
	}

	// Enable Sleep API Rate limit strategy - it will sleep until new tokens are available
	// see https://help.ns1.com/hc/en-us/articles/360020250573-About-API-rate-limiting
	// this strategy would imply the least sleep time for non-parallel client requests
	return &nsone{rest.NewClient(
		http.DefaultClient,
		rest.SetAPIKey(creds["api_token"]),
		func(c *rest.Client) {
			c.RateLimitStrategySleep()
		},
	)}, nil
}

// A wrapper around rest.Client's Zones.Get() implementing retries
// no explicit sleep is needed, it is implemented in NS1 client's RateLimitStrategy we used
func (n *nsone) GetZone(domain string) (*dns.Zone, error) {
	for rtr := 0; ; rtr++ {
		z, httpResp, err := n.Zones.Get(domain, true)
		if httpResp.StatusCode == http.StatusTooManyRequests && rtr < clientRetries {
			continue
		}
		return z, err
	}
}

func (n *nsone) EnsureZoneExists(domain string) error {
	// This enables the create-domains subcommand
	zone := dns.NewZone(domain)

	for rtr := 0; ; rtr++ {
		httpResp, err := n.Zones.Create(zone)
		if err == rest.ErrZoneExists {
			// if domain exists already, just return nil, nothing to do here.
			return nil
		}
		// too many requests - retry w/out waiting. We specified rate limit strategy creating the client
		if httpResp.StatusCode == http.StatusTooManyRequests && rtr < clientRetries {
			continue
		}
		return err
	}
}

func (n *nsone) GetNameservers(domain string) ([]*models.Nameserver, error) {
	var nservers []string

	z, _, err := n.Zones.Get(domain, true)
	if err != nil {
		return nil, err
	}

	// on newly-created domains NS1 may assign nameservers with or without a
	// trailing dot. This is not reflected in the actual DNS records, that
	// always have the trailing dots.
	//
	// Handle both scenarios by stripping dots where existing, before continuing.
	for _, ns := range z.DNSServers {
		if strings.HasSuffix(ns, ".") {
			nservers = append(nservers, ns[0:len(ns)-1])
		} else {
			nservers = append(nservers, ns)
		}
	}
	return models.ToNameservers(nservers)
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (n *nsone) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	z, _, err := n.Zones.Get(domain, true)
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

// GetZoneDNSSEC gets DNSSEC status for zone. Returns true for enabled, false for disabled
// a domain in NS1 can be in 3 states:
//  1. DNSSEC is enabled  (returns true)
//  2. DNSSEC is disabled (returns false)
//  3. some error state   (return false plus the error)
func (n *nsone) GetZoneDNSSEC(domain string) (bool, error) {
	for rtr := 0; ; rtr++ {
		_, httpResp, err := n.DNSSEC.Get(domain)
		// rest.ErrDNSECNotEnabled is our "disabled" state
		if err != nil && err == rest.ErrDNSECNotEnabled {
			return false, nil
		}
		if httpResp.StatusCode == http.StatusTooManyRequests && rtr < clientRetries {
			continue
		}
		// any other errors not expected, let's surface them
		if err != nil {
			return false, err
		}

		// no errors returned, we assume DNSSEC is enabled
		return true, nil
	}
}

// getDomainCorrectionsDNSSEC creates DNSSEC zone corrections based on current state and preference
func (n *nsone) getDomainCorrectionsDNSSEC(domain, toggleDNSSEC string) *models.Correction {

	// get dnssec status from NS1 for domain
	// if errors are returned, we bail out without any DNSSEC corrections
	status, err := n.GetZoneDNSSEC(domain)
	if err != nil {
		return nil
	}

	if toggleDNSSEC == "on" && !status {
		// disabled, but prefer it on, let's enable DNSSEC
		return &models.Correction{
			Msg: "ENABLE DNSSEC",
			F:   func() error { return n.configureDNSSEC(domain, true) },
		}
	} else if toggleDNSSEC == "off" && status {
		// enabled, but prefer it off, let's disable DNSSEC
		return &models.Correction{
			Msg: "DISABLE DNSSEC",
			F:   func() error { return n.configureDNSSEC(domain, false) },
		}
	}
	return nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (n *nsone) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, error) {
	var corrections []*models.Correction
	domain := dc.Name

	// add DNSSEC-related corrections
	if dnssecCorrections := n.getDomainCorrectionsDNSSEC(domain, dc.AutoDNSSEC); dnssecCorrections != nil {
		corrections = append(corrections, dnssecCorrections)
	}

	changes, err := diff2.ByRecordSet(existingRecords, dc, nil)
	if err != nil {
		return nil, err
	}

	for _, change := range changes {
		key := change.Key
		recs := change.New
		desc := strings.Join(change.Msgs, "\n")

		switch change.Type {
		case diff2.REPORT:
			corrections = append(corrections, &models.Correction{Msg: change.MsgsJoined})
		case diff2.CREATE:
			corrections = append(corrections, &models.Correction{
				Msg: desc,
				F:   func() error { return n.add(recs, dc.Name) },
			})
		case diff2.CHANGE:
			corrections = append(corrections, &models.Correction{
				Msg: desc,
				F:   func() error { return n.modify(recs, dc.Name) },
			})
		case diff2.DELETE:
			corrections = append(corrections, &models.Correction{
				Msg: desc,
				F:   func() error { return n.remove(key, dc.Name) },
			})
		default:
			panic(fmt.Sprintf("unhandled inst.Type %s", change.Type))
		}

	}
	return corrections, nil
}

func (n *nsone) add(recs models.Records, domain string) error {
	for rtr := 0; ; rtr++ {
		httpResp, err := n.Records.Create(buildRecord(recs, domain, ""))
		if httpResp.StatusCode == http.StatusTooManyRequests && rtr < clientRetries {
			continue
		}
		return err
	}
}

func (n *nsone) remove(key models.RecordKey, domain string) error {
	if key.Type == "NS1_URLFWD" {
		key.Type = "URLFWD"
	}

	for rtr := 0; ; rtr++ {
		httpResp, err := n.Records.Delete(domain, key.NameFQDN, key.Type)
		if httpResp.StatusCode == http.StatusTooManyRequests && rtr < clientRetries {
			continue
		}
		return err
	}
}

func (n *nsone) modify(recs models.Records, domain string) error {
	for rtr := 0; ; rtr++ {
		httpResp, err := n.Records.Update(buildRecord(recs, domain, ""))
		if httpResp.StatusCode == http.StatusTooManyRequests && rtr < clientRetries {
			continue
		}
		return err
	}
}

// configureDNSSEC configures DNSSEC for a zone. Set 'enabled' to true to enable, false to disable.
// There's a cornercase, in which DNSSEC is globally disabled for the account.
// In that situation, enabling DNSSEC will always fail with:
//
//	#1: ENABLE DNSSEC
//	FAILURE! POST https://api.nsone.net/v1/zones/example.com: 400 DNSSEC support is not enabled for this account. Please contact support@ns1.com to enable it
//
// Unfortunately this is not detectable otherwise, so given that we have a nice error message, we just let this through.
func (n *nsone) configureDNSSEC(domain string, enabled bool) error {
	z, _, err := n.Zones.Get(domain, true)
	if err != nil {
		return err
	}
	z.DNSSEC = &enabled
	for rtr := 0; ; rtr++ {
		httpResp, err := n.Zones.Update(z)
		if httpResp.StatusCode == http.StatusTooManyRequests && rtr < clientRetries {
			continue
		}
		return err
	}
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
			rec.AddAnswer(&dns.Answer{Rdata: strings.Fields(fmt.Sprintf("%d %v", r.MxPreference, r.GetTargetField()))})
		} else if r.Type == "TXT" {
			rec.AddAnswer(&dns.Answer{Rdata: []string{r.GetTargetTXTJoined()}})
		} else if r.Type == "CAA" {
			rec.AddAnswer(&dns.Answer{
				Rdata: []string{
					fmt.Sprintf("%v", r.CaaFlag),
					r.CaaTag,
					r.GetTargetField(),
				}})
		} else if r.Type == "SRV" {
			rec.AddAnswer(&dns.Answer{Rdata: strings.Fields(fmt.Sprintf("%d %d %d %v", r.SrvPriority, r.SrvWeight, r.SrvPort, r.GetTargetField()))})
		} else if r.Type == "NAPTR" {
			rec.AddAnswer(&dns.Answer{Rdata: []string{
				strconv.Itoa(int(r.NaptrOrder)),
				strconv.Itoa(int(r.NaptrPreference)),
				r.NaptrFlags,
				r.NaptrService,
				r.NaptrRegexp,
				r.GetTargetField()}})
		} else if r.Type == "DS" {
			rec.AddAnswer(&dns.Answer{Rdata: []string{
				strconv.Itoa(int(r.DsKeyTag)),
				strconv.Itoa(int(r.DsAlgorithm)),
				strconv.Itoa(int(r.DsDigestType)),
				r.DsDigest}})
		} else if r.Type == "NS1_URLFWD" {
			rec.Type = "URLFWD"
			rec.AddAnswer(&dns.Answer{Rdata: strings.Fields(r.GetTargetField())})
		} else if r.Type == "SVCB" || r.Type == "HTTPS" {
			rec.AddAnswer(&dns.Answer{Rdata: []string{
				strconv.Itoa(int(r.SvcPriority)),
				r.GetTargetField(),
				r.SvcParams}})
		} else if r.Type == "TLSA" {
			rec.AddAnswer(&dns.Answer{Rdata: []string{
				strconv.Itoa(int(r.TlsaUsage)),
				strconv.Itoa(int(r.TlsaSelector)),
				strconv.Itoa(int(r.TlsaMatchingType)),
				r.GetTargetField()}})
		} else {
			rec.AddAnswer(&dns.Answer{Rdata: strings.Fields(r.GetTargetField())})
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
		case "DNSKEY", "RRSIG":
			// if a zone is enabled for DNSSEC, NS1 autoconfigures DNSKEY & RRSIG records.
			// these entries are not modifiable via the API though, so we have to ignore them while converting.
			// 	ie. API returns "405 Operation on DNSSEC record is not allowed" on such operations
			continue
		case "ALIAS":
			rec.Type = rtype
			if err := rec.SetTarget(ans); err != nil {
				return nil, fmt.Errorf("unparsable %s record received from ns1: %w", rtype, err)
			}
		case "URLFWD":
			rec.Type = "NS1_URLFWD"
			if err := rec.SetTarget(ans); err != nil {
				return nil, fmt.Errorf("unparsable %s record received from ns1: %w", rtype, err)
			}
		case "CAA":
			//dnscontrol expects quotes around multivalue CAA entries, API doesn't add them
			xAns := strings.SplitN(ans, " ", 3)
			if err := rec.SetTargetCAAStrings(xAns[0], xAns[1], xAns[2]); err != nil {
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
