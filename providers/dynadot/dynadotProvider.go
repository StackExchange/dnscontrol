package dynadot

import (
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

/*

Dynadot Registrator:

Info required in `creds.json`:
   - key API Key

*/

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanConcur: providers.Cannot(),
}

func init() {
	providers.RegisterRegistrarType("DYNADOT", newDynadot, features)
}

func newDynadot(m map[string]string) (providers.Registrar, error) {
	d := &dynadotProvider{}

	d.key = m["key"]
	if d.key == "" {
		return nil, fmt.Errorf("missing Dynadot key")
	}

	return d, nil
}

func (c *dynadotProvider) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	nss, err := c.getNameservers(dc.Name)
	if err != nil {
		return nil, err
	}
	foundNameservers := strings.Join(nss, ",")

	expected := []string{}
	for _, ns := range dc.Nameservers {
		name := strings.TrimRight(ns.Name, ".")
		expected = append(expected, name)
	}
	sort.Strings(expected)
	expectedNameservers := strings.Join(expected, ",")

	if foundNameservers != expectedNameservers {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Update nameservers (%s) -> (%s)", foundNameservers, expectedNameservers),
				F: func() error {
					return c.updateNameservers(expected, dc.Name)
				},
			},
		}, nil
	}
	return nil, nil
}
