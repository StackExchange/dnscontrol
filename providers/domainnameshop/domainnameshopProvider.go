package domainnameshop

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

var features = providers.DocumentationNotes{}

/*	providers.CanGetZones:            providers.Cannot(),
	providers.CanUseAlias:            providers.Unimplemented(),
	providers.CanUseCAA:              providers.Unimplemented(),
	providers.CanUseDSForChildren:    providers.Unimplemented(),
	providers.CanUsePTR:              providers.Unimplemented(),
	providers.CanUseSRV:              providers.Unimplemented(),
	providers.CanUseSSHFP:            providers.Unimplemented(),
	providers.CanUseTLSA:             providers.Unimplemented(),
	providers.DocCreateDomains:       providers.Unimplemented(),
	providers.DocDualHost:            providers.Unimplemented(),
	providers.DocOfficiallySupported: providers.Unimplemented(),
}*/

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
		RecordAuditor: AuditRecords,
	}

	proxyURL, _ := url.Parse("http://127.0.0.1:8080")
	http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyURL), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	providers.RegisterDomainServiceProviderType("DOMAINNAMESHOP", fns, features)
}

func AuditRecords(records []*models.RecordConfig) error {
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

	differ := diff.New(dc)
	_, create, delete, modify, err := differ.IncrementalDiff(existingRecords)
	if err != nil {
		return nil, err
	}

	var corrections = []*models.Correction{}

	// Delete record
	for _, r := range delete {
		domainID := r.Existing.Original.(*DomainNameShopRecord).DomainID
		recordID := strconv.Itoa(r.Existing.Original.(*DomainNameShopRecord).ID)

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

		dnsR.ID = r.Existing.Original.(*DomainNameShopRecord).ID

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
