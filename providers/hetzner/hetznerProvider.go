package hetzner

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

var features = providers.DocumentationNotes{
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Cannot(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Cannot(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.CanUseTXTMulti:         providers.Cannot(),
}

func init() {
	providers.RegisterDomainServiceProviderType("HETZNER", New, features)
}

// New creates a new API handle.
func New(settings map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	if settings["api_key"] == "" {
		return nil, fmt.Errorf("missing HETZNER api_key")
	}

	api := &hetznerProvider{}

	api.apiKey = settings["api_key"]

	if settings["rate_limited"] == "true" {
		// backwards compatibility
		settings["start_with_default_rate_limit"] = "true"
	}
	if settings["start_with_default_rate_limit"] == "true" {
		api.startRateLimited()
	}

	quota := settings["optimize_for_rate_limit_quota"]
	err := api.requestRateLimiter.setOptimizeForRateLimitQuota(quota)
	if err != nil {
		return nil, fmt.Errorf("unexpected value for optimize_for_rate_limit_quota: %w", err)
	}

	return api, nil
}

// EnsureDomainExists creates the domain if it does not exist.
func (api *hetznerProvider) EnsureDomainExists(domain string) error {
	domains, err := api.ListZones()
	if err != nil {
		return err
	}

	for _, d := range domains {
		if d == domain {
			return nil
		}
	}

	return api.createZone(domain)
}

// GetDomainCorrections returns the corrections for a domain.
func (api *hetznerProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc, err := dc.Copy()
	if err != nil {
		return nil, err
	}

	err = dc.Punycode()
	if err != nil {
		return nil, err
	}
	domain := dc.Name

	// Get existing records
	existingRecords, err := api.GetZoneRecords(domain)
	if err != nil {
		return nil, err
	}

	// Normalize
	models.PostProcessRecords(existingRecords)

	differ := diff.New(dc)
	_, create, del, modify, err := differ.IncrementalDiff(existingRecords)
	if err != nil {
		return nil, err
	}

	var corrections []*models.Correction

	zone, err := api.getZone(domain)
	if err != nil {
		return nil, err
	}

	for _, m := range del {
		record := m.Existing.Original.(*record)
		corr := &models.Correction{
			Msg: m.String(),
			F: func() error {
				return api.deleteRecord(*record)
			},
		}
		corrections = append(corrections, corr)
	}

	var createRecords []record
	createDescription := []string{"Batch creation of records:"}
	for _, m := range create {
		record := fromRecordConfig(m.Desired, zone)
		createRecords = append(createRecords, *record)
		createDescription = append(createDescription, m.String())
	}
	if len(createRecords) > 0 {
		corr := &models.Correction{
			Msg: strings.Join(createDescription, "\n\t"),
			F: func() error {
				return api.bulkCreateRecords(createRecords)
			},
		}
		corrections = append(corrections, corr)
	}

	var modifyRecords []record
	modifyDescription := []string{"Batch modification of records:"}
	for _, m := range modify {
		id := m.Existing.Original.(*record).ID
		record := fromRecordConfig(m.Desired, zone)
		record.ID = id
		modifyRecords = append(modifyRecords, *record)
		modifyDescription = append(modifyDescription, m.String())
	}
	if len(modifyRecords) > 0 {
		corr := &models.Correction{
			Msg: strings.Join(modifyDescription, "\n\t"),
			F: func() error {
				return api.bulkUpdateRecords(modifyRecords)
			},
		}
		corrections = append(corrections, corr)
	}

	return corrections, nil
}

// GetNameservers returns the nameservers for a domain.
func (api *hetznerProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	zone, err := api.getZone(domain)
	if err != nil {
		return nil, err
	}
	nameserver := make([]*models.Nameserver, len(zone.NameServers))
	for i := range zone.NameServers {
		nameserver[i] = &models.Nameserver{Name: zone.NameServers[i]}
	}
	return nameserver, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *hetznerProvider) GetZoneRecords(domain string) (models.Records, error) {
	records, err := api.getAllRecords(domain)
	if err != nil {
		return nil, err
	}
	existingRecords := make([]*models.RecordConfig, len(records))
	for i := range records {
		existingRecords[i] = toRecordConfig(domain, &records[i])
	}
	return existingRecords, nil
}

// ListZones lists the zones on this account.
func (api *hetznerProvider) ListZones() ([]string, error) {
	if err := api.getAllZones(); err != nil {
		return nil, err
	}
	var zones []string
	for i := range api.zones {
		zones = append(zones, i)
	}
	return zones, nil
}
