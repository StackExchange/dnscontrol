package porkbun

func (client *porkbunProvider) ListZones() ([]string, error) {
	zones, err := client.listAllDomains()
	if err != nil {
		return nil, err
	}
	return zones, err
}
