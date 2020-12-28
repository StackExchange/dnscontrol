package msdns

func (c *msdnsProvider) ListZones() ([]string, error) {
	zones, err := c.shell.GetDNSServerZoneAll(c.dnsserver)
	if err != nil {
		return nil, err
	}
	return zones, err
}
