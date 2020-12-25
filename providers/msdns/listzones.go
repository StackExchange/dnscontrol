package activedir

func (c *activedirProvider) ListZones() ([]string, error) {
	zones, err := c.shell.GetDNSServerZoneAll(c.dnsserver)
	if err != nil {
		return nil, err
	}
	return zones, err
}
