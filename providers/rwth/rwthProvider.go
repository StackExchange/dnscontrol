package rwth

import (
	"encoding/json"
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/txtutil"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

// features is used to let dnscontrol know which features are supported by the RWTH DNS Admin.
var features = providers.DocumentationNotes{
	providers.CanAutoDNSSEC:          providers.Unimplemented("Supported by RWTH but not implemented yet."),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseAzureAlias:       providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Unimplemented("DS records are only supported at the apex and require a different API call that hasn't been implemented yet."),
	providers.CanUseNAPTR:            providers.Cannot(),
	providers.CanUsePTR:              providers.Can("PTR records with empty targets are not supported"),
	providers.CanUseSRV:              providers.Can("SRV records with empty targets are not supported."),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

var RWTHDefaultNs = []string{"dns-1.dfn.de", "dns-2.dfn.de", "zs1.rz.rwth-aachen.de", "zs2.rz.rwth-aachen.de"}

// init registers the registrar and the domain service provider with dnscontrol.
func init() {
	fns := providers.DspFuncs{
		Initializer:   New,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("RWTH", fns, features)
}

func New(settings map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	if settings["api_token"] == "" {
		return nil, fmt.Errorf("missing RWTH api_token")
	}

	api := &rwthProvider{apiToken: settings["api_token"]}

	return api, nil
}

func (api *rwthProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
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
	txtutil.SplitSingleLongTxt(dc.Records) // Autosplit long TXT records

	differ := diff.New(dc)
	_, create, del, modify, err := differ.IncrementalDiff(existingRecords)
	if err != nil {
		return nil, err
	}

	var corrections []*models.Correction

	for _, d := range create {
		des := d.Desired
		corrections = append(corrections, &models.Correction{
			Msg: d.String(),
			F:   func() error { return api.createRecord(dc.Name, des) },
		})
	}
	for _, d := range del {
		existingRecord := d.Existing.Original.(RecordReply)
		corrections = append(corrections, &models.Correction{
			Msg: d.String(),
			F:   func() error { return api.destroyRecord(existingRecord) },
		})
	}
	for _, d := range modify {
		rec := d.Desired
		existingID := d.Existing.Original.(RecordReply).ID
		corrections = append(corrections, &models.Correction{
			Msg: d.String(),
			F:   func() error { return api.updateRecord(existingID, *rec) },
		})
	}

	// And deploy if any corrections were applied
	if len(corrections) > 0 {
		corrections = append(corrections, &models.Correction{
			Msg: fmt.Sprintf("Deploy zone %s", domain),
			F:   func() error { return api.deployZone(domain) },
		})
	}

	return corrections, nil
}

// GetNameservers returns the default nameservers for RWTH.
func (api *rwthProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.ToNameservers(RWTHDefaultNs)
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *rwthProvider) GetZoneRecords(domain string) (models.Records, error) {
	records, err := api.getAllRecords(domain)
	if err != nil {
		return nil, err
	}
	foundRecords := models.Records{}
	for i, _ := range records {
		foundRecords = append(foundRecords, &records[i])
	}
	return foundRecords, nil
}

// ListZones lists the zones on this account.
func (api *rwthProvider) ListZones() ([]string, error) {
	if err := api.getAllZones(); err != nil {
		return nil, err
	}
	var zones []string
	for i := range api.zones {
		zones = append(zones, i)
	}
	return zones, nil
}
