package bunnydns

import (
	"encoding/json"
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanAutoDNSSEC:          providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Cannot(),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDHCID:            providers.Cannot(),
	providers.CanUseDS:               providers.Cannot(),
	providers.CanUseDSForChildren:    providers.Cannot(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUseNAPTR:            providers.Cannot(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSOA:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Cannot(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

type bunnydnsProvider struct {
	apiKey string
	zones  map[string]*zone
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   newBunnydns,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("BUNNY_DNS", fns, features)
}

func newBunnydns(settings map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	apiKey := settings["api_key"]
	if apiKey == "" {
		return nil, fmt.Errorf("missing BUNNY_DNS api_key")
	}

	return &bunnydnsProvider{
		apiKey: apiKey,
	}, nil
}

func (b *bunnydnsProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	zone, err := b.findZoneByDomain(domain)
	if err != nil {
		return nil, err
	}

	return models.ToNameservers(zone.Nameservers())
}
