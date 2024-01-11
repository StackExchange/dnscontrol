package porkbun

func (c *porkbunProvider) ListZones() ([]string, error) {
	zones, err := c.listAllDomains()
	if err != nil {
		return nil, err
	}
	return zones, err
}
