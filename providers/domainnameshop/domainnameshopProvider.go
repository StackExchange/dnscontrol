package domainnameshop

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

var features = providers.DocumentationNotes{
	providers.CanAutoDNSSEC:          providers.Cannot(),        // Maybe there is support for it
	providers.CanGetZones:            providers.Unimplemented(), //
	providers.CanUseAlias:            providers.Unimplemented(), // Can possibly be implemented, needs further research
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Unimplemented(), // Seems to support but needs to be implemented
	providers.CanUseDSForChildren:    providers.Unimplemented(), // Seems to support but needs to be implemented
	providers.CanUseNAPTR:            providers.Cannot(),        // Does not seem to support it
	providers.CanUsePTR:              providers.Unimplemented(), // Seems to support but needs to be implemented
	providers.CanUseSOA:              providers.Cannot(),        // Does not seem to support it
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Cannot(),        // Does not seem to support it
	providers.CanUseTLSA:             providers.Unimplemented(), // //Seems to support but needs to be implemented
	providers.DocCreateDomains:       providers.Unimplemented(), // Not tested
	providers.DocDualHost:            providers.Unimplemented(), // Not tested
	providers.DocOfficiallySupported: providers.Cannot(),
}

// dnsimpleProvider is the handle for this provider.
type domainNameShopProvider struct {
	Token  string // The API token
	Secret string // The API secret
}

// Register with the dnscontrol system.
// This establishes the name (all caps), and the function to call to initialize it.
func init() {
	fns := providers.DspFuncs{
		Initializer:   newDomainNameShopProvider,
		RecordAuditor: auditRecords,
	}

	providers.RegisterDomainServiceProviderType("DOMAINNAMESHOP", fns, features)
}

func auditRecords(records []*models.RecordConfig) error {
	return nil
}

// newDomainNameShopProvider creates a DomainNameShop specific DNS provider.
func newDomainNameShopProvider(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	if conf["token"] == "" {
		return nil, fmt.Errorf("no DomainNameShop token provided")
	} else if conf["secret"] == "" {
		return nil, fmt.Errorf("no DomainNameShop secret provided")
	}

	api := &domainNameShopProvider{
		Token:  conf["token"],
		Secret: conf["secret"],
	}

	// Consider testing if creds work
	return api, nil
}

func (api *domainNameShopProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	ns, err := api.getNS(domain)
	if err != nil {
		return nil, err
	}
	return models.ToNameservers(ns)
}

func (api *domainNameShopProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()
	existingRecords, err := api.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	// Normalize
	models.PostProcessRecords(existingRecords)

	// Merge TXT strings to one string
	for _, rc := range dc.Records {
		if rc.HasFormatIdenticalToTXT() {
			rc.SetTargetTXT(strings.Join(rc.TxtStrings, ""))
		}
	}

	// Domainnameshop doesn't allow arbitrary TTLs they must be a multiple of 60.
	for _, record := range dc.Records {
		record.TTL = fixTTL(record.TTL)
	}

	differ := diff.New(dc)
	_, create, delete, modify, err := differ.IncrementalDiff(existingRecords)
	if err != nil {
		return nil, err
	}

	var corrections = []*models.Correction{}

	// Delete record
	for _, r := range delete {
		domainID := r.Existing.Original.(*domainNameShopRecord).DomainID
		recordID := strconv.Itoa(r.Existing.Original.(*domainNameShopRecord).ID)

		corr := &models.Correction{
			Msg: fmt.Sprintf("%s, record id: %s", r.String(), recordID),
			F:   func() error { return api.deleteRecord(domainID, recordID) },
		}
		corrections = append(corrections, corr)
	}

	// Create records
	for _, r := range create {
		domainName := strings.Replace(r.Desired.GetLabelFQDN(), r.Desired.GetLabel()+".", "", -1)
		dnsR, err := api.fromRecordConfig(domainName, r.Desired)
		if err != nil {
			return nil, err
		}
		corr := &models.Correction{
			Msg: r.String(),
			F:   func() error { return api.CreateRecord(domainName, dnsR) },
		}

		corrections = append(corrections, corr)
	}

	for _, r := range modify {
		domainName := strings.Replace(r.Desired.GetLabelFQDN(), r.Desired.GetLabel()+".", "", -1)

		dnsR, err := api.fromRecordConfig(domainName, r.Desired)

		if err != nil {
			return nil, err
		}

		dnsR.ID = r.Existing.Original.(*domainNameShopRecord).ID

		corr := &models.Correction{
			Msg: r.String(),
			F:   func() error { return api.UpdateRecord(dnsR) },
		}

		corrections = append(corrections, corr)
	}

	return corrections, nil
}

func (api *domainNameShopProvider) GetZoneRecords(domain string) (models.Records, error) {
	records, err := api.getDNS(domain)
	if err != nil {
		return nil, err
	}

	var existingRecords []*models.RecordConfig
	for i := range records {
		rC := toRecordConfig(domain, &records[i])
		existingRecords = append(existingRecords, rC)
	}

	return existingRecords, nil
}

const minAllowedTTL = 60
const maxAllowedTTL = 604800
const multiplierTTL = 60

func fixTTL(ttl uint32) uint32 {
	// if the TTL is larger than the largest allowed value, return the largest allowed value
	if ttl > maxAllowedTTL {
		return maxAllowedTTL
	} else if ttl < 60 {
		return minAllowedTTL
	}

	// Return closest rounded down possible

	return (ttl / multiplierTTL) * multiplierTTL
}
