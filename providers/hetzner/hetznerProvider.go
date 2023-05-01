package hetzner

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v3/pkg/txtutil"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

var features = providers.DocumentationNotes{
	providers.CanAutoDNSSEC:          providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Can(),
	providers.CanUseDSForChildren:    providers.Cannot(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUseNAPTR:            providers.Cannot(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSOA:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Cannot(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   New,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("HETZNER", fns, features)
}

// New creates a new API handle.
func New(settings map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	apiKey := settings["api_key"]
	if apiKey == "" {
		return nil, fmt.Errorf("missing HETZNER api_key")
	}

	return &hetznerProvider{
		apiKey: apiKey,
	}, nil
}

// EnsureZoneExists creates a zone if it does not exist
func (api *hetznerProvider) EnsureZoneExists(domain string) error {
	domains, err := api.ListZones()
	if err != nil {
		return err
	}

	for _, d := range domains {
		if d == domain {
			return nil
		}
	}

	// reset zone cache
	api.zones = nil
	return api.createZone(domain)
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (api *hetznerProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, error) {
	domain := dc.Name

	txtutil.SplitSingleLongTxt(dc.Records) // Autosplit long TXT records

	var corrections []*models.Correction
	var differ diff.Differ
	if !diff2.EnableDiff2 {
		differ = diff.New(dc)
	} else {
		differ = diff.NewCompat(dc)
	}
	_, create, del, modify, err := differ.IncrementalDiff(existingRecords)
	if err != nil {
		return nil, err
	}

	z, err := api.getZone(domain)
	if err != nil {
		return nil, err
	}

	corrections = make([]*models.Correction, 0, len(del)+1+1)
	for _, m := range del {
		r := m.Existing.Original.(*record)
		corr := &models.Correction{
			Msg: m.String(),
			F: func() error {
				return api.deleteRecord(r)
			},
		}
		corrections = append(corrections, corr)
	}

	createRecords := make([]record, len(create))
	createDescription := make([]string, len(create)+1)
	createDescription[0] = "Batch creation of records:"
	for i, m := range create {
		createRecords[i] = fromRecordConfig(m.Desired, z)
		createDescription[i+1] = m.String()
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

	modifyRecords := make([]record, len(modify))
	modifyDescription := make([]string, len(modify)+1)
	modifyDescription[0] = "Batch modification of records:"
	for i, m := range modify {
		r := fromRecordConfig(m.Desired, z)
		r.ID = m.Existing.Original.(*record).ID
		modifyRecords[i] = r
		modifyDescription[i+1] = m.String()
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
	z, err := api.getZone(domain)
	if err != nil {
		return nil, err
	}
	nameserver := make([]*models.Nameserver, len(z.NameServers))
	for i, s := range z.NameServers {
		nameserver[i] = &models.Nameserver{Name: s}
	}
	return nameserver, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *hetznerProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	records, err := api.getAllRecords(domain)
	if err != nil {
		return nil, err
	}
	existingRecords := make([]*models.RecordConfig, len(records))
	for i := range records {
		existingRecords[i], err = toRecordConfig(domain, &records[i])
		if err != nil {
			return nil, err
		}
	}
	return existingRecords, nil
}

// ListZones lists the zones on this account.
func (api *hetznerProvider) ListZones() ([]string, error) {
	if err := api.getAllZones(); err != nil {
		return nil, err
	}
	zones := make([]string, 0, len(api.zones))
	for domain := range api.zones {
		zones = append(zones, domain)
	}
	return zones, nil
}
