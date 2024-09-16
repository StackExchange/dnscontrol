package dnsmadeeasy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Cannot(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Cannot(),
	providers.CanUseDSForChildren:    providers.Cannot(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Cannot(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can("System NS records cannot be edited. Custom apex NS records can be added/changed/deleted."),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	const providerName = "DNSMADEEASY"
	const providerMaintainer = "@vojtad"
	fns := providers.DspFuncs{
		Initializer:   New,
		RecordAuditor: AuditRecords,
	}

	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	providers.RegisterMaintainer(providerName, providerMaintainer)
}

// New creates a new API handle.
func New(settings map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	if settings["api_key"] == "" {
		return nil, fmt.Errorf("missing DNSMADEEASY api_key")
	}

	if settings["secret_key"] == "" {
		return nil, fmt.Errorf("missing DNSMADEEASY secret_key")
	}

	sandbox := false
	if settings["sandbox"] != "" {
		sandbox = true
	}

	debug := false
	if os.Getenv("DNSMADEEASY_DEBUG_HTTP") == "1" {
		debug = true
	}

	api := newProvider(settings["api_key"], settings["secret_key"], sandbox, debug)

	return api, nil
}

// // GetDomainCorrections returns the corrections for a domain.
// func (api *dnsMadeEasyProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
// 	dc, err := dc.Copy()
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = dc.Punycode()
// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, rec := range dc.Records {
// 		if rec.Type == "ALIAS" {
// 			// ALIAS is called ANAME on DNS Made Easy
// 			rec.Type = "ANAME"
// 		} else if rec.Type == "NS" {
// 			// NS records have fixed TTL on DNS Made Easy and it cannot be changed
// 			rec.TTL = fixedNameServerRecordTTL
// 		}
// 	}

// 	domainName := dc.Name

// 	// Get existing records
// 	existingRecords, err := api.GetZoneRecords(domainName)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Normalize
// 	models.PostProcessRecords(existingRecords)

// 	return api.GetZoneRecordsCorrections(dc, existingRecords)
// }

func (api *dnsMadeEasyProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, int, error) {
	domainName := dc.Name
	domain, err := api.findDomain(domainName)
	if err != nil {
		return nil, 0, err
	}

	for _, rec := range dc.Records {
		if rec.Type == "ALIAS" {
			// ALIAS is called ANAME on DNS Made Easy
			rec.Type = "ANAME"
		} else if rec.Type == "NS" {
			// NS records have fixed TTL on DNS Made Easy and it cannot be changed
			rec.TTL = fixedNameServerRecordTTL
		}
	}

	toReport, create, del, modify, actualChangeCount, err := diff.NewCompat(dc).IncrementalDiff(existingRecords)
	if err != nil {
		return nil, 0, err
	}
	// Start corrections with the reports
	corrections := diff.GenerateMessageCorrections(toReport)

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

	return corrections, actualChangeCount, nil
}

// EnsureZoneExists creates a zone if it does not exist
func (api *dnsMadeEasyProvider) EnsureZoneExists(domain string) error {
	exists, err := api.domainExists(domain)
	if err != nil {
		return err
	}

	// domain already exists
	if exists {
		return nil
	}

	return api.createDomain(domain)
}

// GetNameservers returns the nameservers for a domain.
func (api *dnsMadeEasyProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	nameServers, err := api.fetchDomainNameServers(domain)
	if err != nil {
		return nil, err
	}

	return models.ToNameservers(nameServers)
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *dnsMadeEasyProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	records, err := api.fetchDomainRecords(domain)
	if err != nil {
		return nil, err
	}

	nameServers, err := api.fetchDomainNameServers(domain)
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

	for i := range nameServers {
		existingRecords = append(existingRecords, systemNameServerToRecordConfig(domain, nameServers[i]))
	}

	return existingRecords, nil
}

// ListZones lists the zones on this account.
func (api *dnsMadeEasyProvider) ListZones() ([]string, error) {
	if err := api.loadDomains(); err != nil {
		return nil, err
	}

	var zones []string
	for i := range api.domains {
		zones = append(zones, i)
	}

	return zones, nil
}
