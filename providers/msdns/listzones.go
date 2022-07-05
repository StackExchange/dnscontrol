package msdns

import "github.com/StackExchange/dnscontrol/v3/internal/dnscontrol"

func (client *msdnsProvider) ListZones(_ dnscontrol.Context) ([]string, error) {
	zones, err := client.shell.GetDNSServerZoneAll(client.dnsserver)
	if err != nil {
		return nil, err
	}
	return zones, err
}
