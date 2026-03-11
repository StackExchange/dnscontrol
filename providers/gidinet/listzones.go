package gidinet

// ListZones returns all the zones in the account.
func (c *gidinetProvider) ListZones() ([]string, error) {
	domains, err := c.domainGetList()
	if err != nil {
		return nil, err
	}

	var zones []string
	for _, domain := range domains {
		// Combine domain name and extension to get the full zone name
		zoneName := domain.DomainName + "." + domain.DomainExtension
		zones = append(zones, zoneName)
	}

	return zones, nil
}
