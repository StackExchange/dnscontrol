package misaka

import (
	"fmt"

	"github.com/DNSControl/dnscontrol/v5/models"
)

// ListZones returns every zone the API token can see.
func (p *misakaProvider) ListZones() ([]string, error) {
	zones, err := p.client.listZones()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(zones))
	for _, z := range zones {
		names = append(names, z.Name)
	}

	return names, nil
}

// EnsureZoneExists creates the zone if the account does not have it yet.
// create-domains calls this without listing the zones first, so it checks
// for itself.
func (p *misakaProvider) EnsureZoneExists(dc *models.DomainConfig) error {
	err := p.client.getZoneSettings(dc.Name)
	if err == nil {
		return nil
	}
	if !isNotFound(err) {
		return err
	}

	if err := p.client.createZone(dc.Name); err != nil {
		return fmt.Errorf("creating zone %s: %w", dc.Name, err)
	}
	return nil
}
