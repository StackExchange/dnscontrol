package easyname

import (
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

type easynameProvider struct {
	apikey   string
	apiauth  string
	signSalt string
	domains  map[string]easynameDomain
}

func init() {
	providers.RegisterRegistrarType("EASYNAME", newEasyname)
}

func newEasyname(m map[string]string) (providers.Registrar, error) {
	api := &easynameProvider{}

	if m["email"] == "" || m["userid"] == "" || m["apikey"] == "" || m["authsalt"] == "" || m["signsalt"] == "" {
		return nil, printer.Errorf("missing easyname email, userid, apikey, authsalt and/or signsalt")
	}

	api.apikey, api.signSalt = m["apikey"], m["signsalt"]
	composed := fmt.Sprintf(m["authsalt"], m["userid"], m["email"])
	api.apiauth = hashEncodeString(composed)

	return api, nil
}

// GetRegistrarCorrections gathers corrections that would being n to match dc.
func (c *easynameProvider) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	domain, err := c.getDomain(dc.Name)
	if err != nil {
		return nil, err
	}

	nss := []string{}
	for _, ns := range []string{domain.NameServer1, domain.NameServer2, domain.NameServer3, domain.NameServer4, domain.NameServer5, domain.NameServer6} {
		if ns != "" {
			nss = append(nss, ns)
		}
	}
	sort.Strings(nss)
	foundNameservers := strings.Join(nss, ",")

	expected := []string{}
	for _, ns := range dc.Nameservers {
		expected = append(expected, ns.Name)
	}
	sort.Strings(expected)
	expectedNameservers := strings.Join(expected, ",")

	if foundNameservers != expectedNameservers {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Update nameservers %s -> %s", foundNameservers, expectedNameservers),
				F: func() error {
					return c.updateNameservers(expected, domain.ID)
				},
			},
		}, nil
	}
	return nil, nil
}
