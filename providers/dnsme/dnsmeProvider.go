package dnsme

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/txtutil"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.CanUseSSHFP:            providers.Cannot(),
	providers.CanUseDS:               providers.Cannot(),
	providers.CanUseDSForChildren:    providers.Cannot(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   New,
		RecordAuditor: AuditRecords,
	}

	providers.RegisterDomainServiceProviderType("DNSME", fns, features)
}

// New creates a new API handle.
func New(settings map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	if settings["api_key"] == "" {
		return nil, fmt.Errorf("missing DNSME api_key")
	}

	if settings["secret_key"] == "" {
		return nil, fmt.Errorf("missing DNSME secret_key")
	}

	sandbox := false
	if settings["sandbox"] != "" {
		sandbox = true
	}

	debug := false
	if os.Getenv("DNSME_DEBUG") == "1" {
		debug = true
	}

	api := newDnsmeProvider(settings["api_key"], settings["secret_key"], sandbox, debug)

	return api, nil
}

// GetDomainCorrections returns the corrections for a domain.
func (api *dnsmeProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc, err := dc.Copy()
	if err != nil {
		return nil, err
	}

	err = dc.Punycode()
	if err != nil {
		return nil, err
	}

	dc.Filter(func(r *models.RecordConfig) bool {
		// Filter out apex NS records, DNS Made Easy doesn't support changing them
		if r.Type == "NS" && r.GetLabel() == "@" {
			return false
		}

		// Filter out custom HTTPRED records, changing them is not supported
		if r.Type == "HTTPRED" {
			return false
		}

		return true
	})

	// ALIAS is called ANAME on DNS Made Easy
	for _, rec := range dc.Records {
		if rec.Type == "ALIAS" {
			rec.Type = "ANAME"
		}
	}

	domainName := dc.Name

	domain, err := api.findDomain(domainName)
	if err != nil {
		return nil, err
	}

	// Get existing records
	existingRecords, err := api.GetZoneRecords(domainName)
	if err != nil {
		return nil, err
	}

	// Normalize
	models.PostProcessRecords(existingRecords)
	txtutil.SplitSingleLongTxt(dc.Records) // Autosplit long TXT records

	differ := diff.New(dc)
	_, create, del, modify, err := differ.IncrementalDiff(existingRecords)
	if err != nil {
		return nil, err
	}

	var corrections []*models.Correction

	var deleteRecordIds []int
	deleteDescription := []string{"Batch deletion of records:"}
	for _, m := range del {
		originalRecordID := m.Existing.Original.(*recordResponseDataEntry).ID
		deleteRecordIds = append(deleteRecordIds, originalRecordID)
		deleteDescription = append(deleteDescription, m.String())
	}

	if len(deleteRecordIds) > 0 {
		corr := &models.Correction{
			Msg: strings.Join(deleteDescription, "\n\t"),
			F: func() error {
				return api.deleteRecords(domain.ID, deleteRecordIds)
			},
		}
		corrections = append(corrections, corr)
	}

	var createRecords []recordRequestData
	createDescription := []string{"Batch creation of records:"}
	for _, m := range create {
		record := fromRecordConfig(m.Desired)
		createRecords = append(createRecords, *record)
		createDescription = append(createDescription, m.String())
	}

	if len(createRecords) > 0 {
		corr := &models.Correction{
			Msg: strings.Join(createDescription, "\n\t"),
			F: func() error {
				return api.createRecords(domain.ID, createRecords)
			},
		}
		corrections = append(corrections, corr)
	}

	var modifyRecords []recordRequestData
	modifyDescription := []string{"Batch modification of records:"}
	for _, m := range modify {
		originalRecord := m.Existing.Original.(*recordResponseDataEntry)

		record := fromRecordConfig(m.Desired)
		record.ID = originalRecord.ID
		record.GtdLocation = originalRecord.GtdLocation

		modifyRecords = append(modifyRecords, *record)
		modifyDescription = append(modifyDescription, m.String())
	}

	if len(modifyRecords) > 0 {
		corr := &models.Correction{
			Msg: strings.Join(modifyDescription, "\n\t"),
			F: func() error {
				return api.updateRecords(domain.ID, modifyRecords)
			},
		}
		corrections = append(corrections, corr)
	}

	return corrections, nil
}

// EnsureDomainExists returns an error if domain doesn't exist.
func (api *dnsmeProvider) EnsureDomainExists(domainName string) error {
	exists, err := api.domainExists(domainName)
	if err != nil {
		return err
	}

	// domain already exists
	if exists {
		return nil
	}

	return api.createDomain(domainName)
}

// GetNameservers returns the nameservers for a domain.
func (api *dnsmeProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	nameServers, err := api.fetchDomainNameServers(domain)
	if err != nil {
		return nil, err
	}

	return models.ToNameservers(nameServers)
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *dnsmeProvider) GetZoneRecords(domain string) (models.Records, error) {
	records, err := api.fetchDomainRecords(domain)
	if err != nil {
		return nil, err
	}

	existingRecords := make([]*models.RecordConfig, 0, len(records))
	for i := range records {
		// Ignore HTTPRED and SPF records
		if records[i].Type == "HTTPRED" || records[i].Type == "SPF" {
			continue
		}
		existingRecords = append(existingRecords, toRecordConfig(domain, &records[i]))
	}

	return existingRecords, nil
}

// ListZones lists the zones on this account.
func (api *dnsmeProvider) ListZones() ([]string, error) {
	if err := api.loadDomains(); err != nil {
		return nil, err
	}

	var zones []string
	for i := range api.domains {
		zones = append(zones, i)
	}

	return zones, nil
}
