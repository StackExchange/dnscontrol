package cscglobal

import (
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/internal/dnscontrol"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// GetRegistrarCorrections gathers corrections that would being n to match dc.
func (client *providerClient) GetRegistrarCorrections(_ dnscontrol.Context, dc *models.DomainConfig) ([]*models.Correction, error) {
	nss, err := client.getNameservers(dc.Name)
	if err != nil {
		return nil, err
	}
	foundNameservers := strings.Join(nss, ",")

	expected := []string{}
	for _, ns := range dc.Nameservers {
		if ns.Name[len(ns.Name)-1] == '.' {
			// When this code was written ns.Name never included a single trailing dot.
			// If that changes, the code should change too.
			return nil, fmt.Errorf("name server includes a trailing dot, has the API changed?")
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
					return client.updateNameservers(expected, dc.Name)
				},
			},
		}, nil
	}
	return nil, nil
}
