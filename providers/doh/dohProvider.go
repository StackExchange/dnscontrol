package doh

import (
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

/*

DNS over HTTPS 'Registrar':

Info required in `creds.json`:
   - host                DNS over HTTPS host (eg 9.9.9.9)
*/

func init() {
	providers.RegisterRegistrarType("DNSOVERHTTPS", newDNSOverHTTPS)
}

func newDNSOverHTTPS(m map[string]string) (providers.Registrar, error) {
	api := &dohProvider{
		host: m["host"],
	}
	if api.host == "" {
		api.host = "dns.google"
	}
	return api, nil
}

// GetRegistrarCorrections gathers corrections that would being n to match dc.
func (c *dohProvider) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	nss, err := c.getNameservers(dc.Name)
	if err != nil {
		return nil, err
	}
	foundNameservers := strings.Join(nss, ",")

	expected := []string{}
	for _, ns := range dc.Nameservers {
		expected = append(expected, ns.Name)
	}
	sort.Strings(expected)
	expectedNameservers := strings.Join(expected, ",")

	if foundNameservers == expectedNameservers {
		return nil, nil
	}

	return []*models.Correction{
		{
			Msg: fmt.Sprintf("Update nameservers %s -> %s", foundNameservers, expectedNameservers),
			F: func() error {
				return c.updateNameservers(expected, dc.Name)
			},
		},
	}, nil
}
