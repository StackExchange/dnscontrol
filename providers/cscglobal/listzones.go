package cscglobal

import "github.com/StackExchange/dnscontrol/v3/internal/dnscontrol"

// ListZones returns all the zones in an account
func (client *providerClient) ListZones(_ dnscontrol.Context) ([]string, error) {

	return client.getDomains()
}
