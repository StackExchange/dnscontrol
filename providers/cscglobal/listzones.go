package cscglobal

// ListZones returns all the zones in an account
func (client *providerClient) ListZones() ([]string, error) {

	return client.getZones()
}

// ListDomains returns all the zones in an account
func (client *providerClient) ListDomains() ([]string, error) {

	return client.getDomains()
}
