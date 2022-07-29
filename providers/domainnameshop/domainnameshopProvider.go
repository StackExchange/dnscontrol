package domainnameshop

import (
	"encoding/json"
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

/**

DomainNameShop Provider

Info required in 'creds.json':
	- token		API Token
	- secret	API Secret

*/

type domainNameShopProvider struct {
	Token  string // The API token
	Secret string // The API secret
}

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
	providers.CanUseTLSA:             providers.Unimplemented(), // Seems to support but needs to be implemented
	providers.DocCreateDomains:       providers.Unimplemented(), // Not tested
	providers.DocDualHost:            providers.Unimplemented(), // Not tested
	providers.DocOfficiallySupported: providers.Cannot(),
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

func auditRecords(records []*models.RecordConfig) error {
	return nil
}

type domainResponse struct {
	ID          int      `json:"id"`
	Domain      string   `json:"domain"`
	Nameservers []string `json:"nameservers"`
}

type domainNameShopRecord struct {
	ID             int    `json:"id"`
	Host           string `json:"host"`
	TTL            uint16 `json:"ttl,omitempty"`
	Type           string `json:"type"`
	Data           string `json:"data"`
	Priority       string `json:"priority,omitempty"`
	ActualPriority uint16
	Weight         string `json:"weight,omitempty"`
	ActualWeight   uint16
	Port           string `json:"port,omitempty"`
	ActualPort     uint16
	CAATag         string `json:"tag,omitempty"`
	ActualCAAFlag  string `json:"flags,omitempty"`
	CAAFlag        uint64
	DomainID       string
}
