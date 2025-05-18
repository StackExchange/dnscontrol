package infomaniak

import (
	"encoding/json"
	"errors"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

// infomaniakProvider is the handle for operations.
type infomaniakProvider struct {
	apiToken string // the account access token
}

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanGetZones: providers.Can(),
}

func newInfomaniak(m map[string]string, message json.RawMessage) (providers.DNSServiceProvider, error) {
	api := &infomaniakProvider{}
	api.apiToken = m["token"]
	if api.apiToken == "" {
		return nil, errors.New("missing Infomaniak personal access token")
	}

	return api, nil
}

func init() {
	const providerName = "INFOMANIAK"
	const providerMaintainer = "@jbelien"
	fns := providers.DspFuncs{
		Initializer: newInfomaniak,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	providers.RegisterMaintainer(providerName, providerMaintainer)
}

func (p *infomaniakProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	zone, err := p.getDNSZone(domain)
	if err != nil {
		return nil, err
	}

	return models.ToNameservers(zone.Nameservers)
}

func (p *infomaniakProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	panic("GetZoneRecords not implemented")
}

func (p *infomaniakProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, int, error) {
	panic("GetZoneRecordsCorrections not implemented")
}
