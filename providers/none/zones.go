package none

func (handle none) ListZones() ([]string, error) {
	return nil, nil
}

func (handle none) EnsureZoneExists(domain string, metadata map[string]string) error {
	return nil
}
