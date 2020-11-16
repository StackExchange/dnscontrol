package cscglobal

import (
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

/*

CSC Global Registrar:

Info required in `creds.json`:
   - api-key             Api Key
   - user-token          User Token
   - notification_emails (optional) Comma separated list of email addresses to send notifications to
*/

func init() {
	providers.RegisterRegistrarType("CSCGLOBAL", newCscGlobal)
}

func newCscGlobal(m map[string]string) (providers.Registrar, error) {
	api := &cscglobalProvider{}

	api.key, api.token = m["api-key"], m["user-token"]
	if api.key == "" || api.token == "" {
		return nil, fmt.Errorf("missing CSC Global api-key and/or user-token")
	}

	if m["notification_emails"] != "" {
		api.notifyEmails = strings.Split(m["notification_emails"], ",")
	}

	return api, nil
}

// GetRegistrarCorrections gathers corrections that would being n to match dc.
func (c *cscglobalProvider) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	nss, err := c.getNameservers(dc.Name)
	if err != nil {
		return nil, err
	}
	foundNameservers := strings.Join(nss, ",")

	expected := []string{}
	for _, ns := range dc.Nameservers {
		if ns.Name[len(ns.Name)-1] == '.' {
			// When this code was written ns.Name never included a single trailing dot.
			// If that changes, the code should change too.
			return nil, fmt.Errorf("Name server includes a trailing dot, has the API changed?")
		}
		expected = append(expected, ns.Name)
	}
	sort.Strings(expected)
	expectedNameservers := strings.Join(expected, ",")

	if foundNameservers != expectedNameservers {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Update nameservers %s -> %s", foundNameservers, expectedNameservers),
				F: func() error {
					return c.updateNameservers(expected, dc.Name)
				},
			},
		}, nil
	}
	return nil, nil
}
