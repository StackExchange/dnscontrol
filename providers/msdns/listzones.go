package msdns

func (client *msdnsProvider) ListZones() ([]string, error) {
	zones, err := client.shell.GetDNSServerZoneAll(client.dnsserver)
	if err != nil {
		return nil, err
	}
	return zones, err
}
