package ns1

import (
	"encoding/json"

	"fmt"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
)

func init() {
	providers.RegisterDomainServiceProviderType("NS1", newProvider)
}

type nsone struct {
	token string
}

func newProvider(creds map[string]string, meta json.RawMessage) (providers.DNSServiceProvider, error) {
	if creds["api_token"] == "" {
		return nil, fmt.Errorf("api_token required for ns1")
	}
	return &nsone{
		token: creds["api_token"],
	}, nil
}

func (n *nsone) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return nil, nil
}
func (n *nsone) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	return nil, nil
}
