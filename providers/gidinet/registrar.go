package gidinet

import (
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
)

// GetRegistrarCorrections returns corrections to update domain nameserver delegation
func (c *gidinetProvider) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	// Get current nameservers from registrar
	existing, err := c.getNameserversForDomain(dc.Name)
	if err != nil {
		return nil, err
	}

	// Normalize existing nameservers (lowercase, sorted)
	for i, ns := range existing {
		existing[i] = strings.ToLower(strings.TrimSuffix(ns, "."))
	}
	sort.Strings(existing)
	existingStr := strings.Join(existing, ",")

	// Get desired nameservers from config
	desired := models.NameserversToStrings(dc.Nameservers)
	// Normalize desired nameservers (lowercase, no trailing dot, sorted)
	for i, ns := range desired {
		desired[i] = strings.ToLower(strings.TrimSuffix(ns, "."))
	}
	// Deduplicate nameservers (can happen when NAMESERVER() and DNS provider both add them)
	seen := make(map[string]bool)
	var uniqueDesired []string
	for _, ns := range desired {
		if !seen[ns] {
			seen[ns] = true
			uniqueDesired = append(uniqueDesired, ns)
		}
	}
	desired = uniqueDesired
	sort.Strings(desired)
	desiredStr := strings.Join(desired, ",")

	// Compare and return correction if different
	if existingStr != desiredStr {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Update nameservers from [%s] to [%s]", existingStr, desiredStr),
				F: func() error {
					return c.setNameservers(dc.Name, desired)
				},
			},
		}, nil
	}

	return nil, nil
}
