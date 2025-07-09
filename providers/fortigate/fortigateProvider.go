package fortigate

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

// Feature Declaration

var features = providers.DocumentationNotes{
	providers.CanGetZones:            providers.Can(),
	providers.CanUsePTR:              providers.Cannot(), // FortiGate does not really support ARPA Zones and handles PTR records really weired
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanConcur:              providers.Unimplemented(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(), // unofficial integration
}

// Provider Registration

func init() {
	const providerName = "FORTIGATE"
        const providerMaintainer = "@KlettIT"
	providers.RegisterDomainServiceProviderType(providerName, providers.DspFuncs{
		Initializer:   NewFortiGate,
		RecordAuditor: AuditRecords,
	}, features)
        providers.RegisterMaintainer(providerName, providerMaintainer)
}

// Provider Struct

type fortigateProvider struct {
	vdom     string
	host     string
	apiKey   string
	insecure bool
	client   *apiClient
}

// Constructor

func NewFortiGate(m map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	host, vdom, apiKey := m["host"], m["vdom"], m["apiKey"]

	var missing []string
	if host == "" {
		missing = append(missing, "host")
	}
	if vdom == "" {
		missing = append(missing, "vdom")
	}
	if apiKey == "" {
		missing = append(missing, "apiKey")
	}
	if len(missing) > 0 {
		return nil, errors.New("Fortigate provider: missing required field(s): " + strings.Join(missing, ", "))
	}

	insecure := strings.EqualFold(m["insecure_tls"], "true")

	p := &fortigateProvider{
		host:     host,
		vdom:     vdom,
		apiKey:   apiKey,
		insecure: insecure,
	}
	p.client = newClient(host, vdom, apiKey, insecure)
	return p, nil
}

// Record Fetching

func (p *fortigateProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	records := models.Records{}

	// Request the zone object from FortiGate
	path := fmt.Sprintf("system/dns-database/%s", strings.TrimSuffix(domain, "."))

	// According to the API, "results" is an array of objects
	var resp struct {
		Results []struct {
			DNSEntry []fgDNSRecord `json:"dns-entry"`
		} `json:"results"`
	}

	err := p.client.do("GET", path, nil, nil, &resp)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			// Zone does not exist yet – return empty record list
			return records, nil
		}
		return nil, fmt.Errorf("fortigate: fetching zone %q failed: %w", domain, err)
	}

	if len(resp.Results) == 0 {
		// Zone exists but no dns-entry data found
		return records, nil
	}

	// Convert native records to dnscontrol Records
	for _, n := range resp.Results[0].DNSEntry {
		rc, err := nativeToRecord(domain, n)
		if err != nil {
			return nil, err
		}
		records = append(records, rc)
	}

	return records, nil
}

// Correction Planning

func (p *fortigateProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, int, error) {

	domain := dc.Name

	var corrections []*models.Correction

	result, err := diff2.ByZone(existingRecords, dc, nil)
	if err != nil {
		return nil, 0, err
	}

	msgs, changed, actualChangeCount := result.Msgs, result.HasChanges, result.ActualChangeCount

	if changed {
		msgs = append(msgs, "Zone update for "+domain)
		msg := strings.Join(msgs, "\n")

		resourceRecords, errs := recordsToNative(result.DesiredPlus)

		if len(errs) > 0 {
			return nil, 0, fmt.Errorf("failed to convert records: %v", errs)
		}

		if resourceRecords == nil {
			resourceRecords = []*fgDNSRecord{}
		}

		payload := map[string]any{
			"forwarder": nil,
			"dns-entry": resourceRecords,
		}

		corrections = append(corrections,
			&models.Correction{
				Msg: msg,
				F: func() error {

					if err := p.EnsureZoneExists(dc.Name); err != nil {
						return err
					}

					return p.client.do("PUT", "system/dns-database/"+dc.Name, nil, payload, nil)
				},
			})
	}

	return corrections, actualChangeCount, nil
}

// Zone Existence Check & Creation
func (p *fortigateProvider) EnsureZoneExists(domain string) error {
	var probe struct{ Results []any }

	err := p.client.do("GET", "system/dns-database/"+domain, nil, nil, &probe)
	switch {
	case err == nil && len(probe.Results) > 0:
		return nil // already exists

	case isNotFound(err):
		body := map[string]any{"name": domain, "domain": domain, "forwarder": nil}
		return p.client.do("POST", "system/dns-database", nil, body, nil)

	default:
		return err
	}
}

// Misc DNSControl Plumbing

func (p *fortigateProvider) GetNameservers(string) ([]*models.Nameserver, error) {
	return nil, nil // FortiGate is authoritative only internally
}

func (p *fortigateProvider) ListZones() ([]string, error) {
	var resp struct {
		Results []struct{ Name string } `json:"results"`
	}
	if err := p.client.do("GET", "system/dns-database", nil, nil, &resp); err != nil {
		return nil, err
	}
	zones := make([]string, len(resp.Results))
	for i, z := range resp.Results {
		zones[i] = z.Name
	}
	return zones, nil
}
