package netcup

import (
	"encoding/json"
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

var features = providers.DocumentationNotes{
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseTXTMulti:         providers.Can(),
	providers.CanGetZones:            providers.Cannot(),
}

func init() {
	providers.RegisterDomainServiceProviderType("NETCUP", New, features)
}

func New(settings map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	if settings["api-key"] == "" || settings["api-password"] == "" || settings["customer-number"] == "" {
		return nil, fmt.Errorf("missing netcup login parameters")
	}

	api := &api{}
	err := api.login(settings["api-key"], settings["api-password"], settings["customer-number"])
	if err != nil {
		return nil, fmt.Errorf("login to netcup DNS failed, please check your credentials: %v", err)
	}
	return api, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *api) GetZoneRecords(domain string) (models.Records, error) {
	records, err := api.getRecords(domain)
	if err != nil {
		return nil, err
	}
	existingRecords := make([]*models.RecordConfig, len(records))
	for i := range records {
		existingRecords[i] = toRecordConfig(domain, &records[i])
	}
	return existingRecords, nil
}

// GetNameservers returns the nameservers for a domain.
// As netcup doesn't support setting nameservers over this API, these are static.
// Domains not managed by netcup DNS will return an error
func (api *api) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.ToNameservers([]string{
		"root-dns.netcup.net",
		"second-dns.netcup.net",
		"third-dns.netcup.net",
	})
}

// GetDomainCorrections returns the corrections for a domain.
func (api *api) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc, err := dc.Copy()
	if err != nil {
		return nil, err
	}

	dc.Punycode()
	domain := dc.Name

	// Setting the TTL is not supported for netcup
	for _, r := range dc.Records {
		r.TTL = 0
	}

	// Filter out types we can't modify (like NS)
	newRecords := models.Records{}
	for _, r := range dc.Records {
		if r.Type != "NS" {
			newRecords = append(newRecords, r)
		}
	}
	dc.Records = newRecords

	// Check existing set
	existingRecords, err := api.GetZoneRecords(domain)
	if err != nil {
		return nil, err
	}

	// Normalize
	models.PostProcessRecords(existingRecords)
	differ := diff.New(dc)
	_, create, del, modify := differ.IncrementalDiff(existingRecords)

	var corrections []*models.Correction

	// Deletes first so changing type works etc.
	for _, m := range del {
		req := m.Existing.Original.(*record)
		corr := &models.Correction{
			Msg: fmt.Sprintf("%s, Netcup ID: %s", m.String(), req.Id),
			F: func() error {
				return api.deleteRecord(domain, req)
			},
		}
		corrections = append(corrections, corr)
	}

	for _, m := range create {
		req := fromRecordConfig(m.Desired)
		corr := &models.Correction{
			Msg: m.String(),
			F: func() error {
				return api.createRecord(domain, req)
			},
		}
		corrections = append(corrections, corr)
	}
	for _, m := range modify {
		id := m.Existing.Original.(*record).Id
		req := fromRecordConfig(m.Desired)
		req.Id = id
		corr := &models.Correction{
			Msg: fmt.Sprintf("%s, Netcup ID: %s: ", m.String(), id),
			F: func() error {
				return api.modifyRecord(domain, req)
			},
		}
		corrections = append(corrections, corr)
	}

	return corrections, nil
}
