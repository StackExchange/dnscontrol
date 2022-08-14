package cscglobal

// ListZones returns all the zones in an account
func (client *providerClient) ListZones() ([]string, error) {

	return client.getDomains()
}
