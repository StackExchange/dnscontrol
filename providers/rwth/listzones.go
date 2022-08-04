package rwth

// ListZones lists the zones on this account.
func (api *rwthProvider) ListZones() ([]string, error) {
	if err := api.getAllZones(); err != nil {
		return nil, err
	}
	var zones []string
	for i := range api.zones {
		zones = append(zones, i)
	}
	return zones, nil
}
