package none

// ListZones returns a list of zones that the None provider manages.
// Since the None provider does not manage any zones, it returns an empty list.
func (n None) ListZones() ([]string, error) {
	return nil, nil
}
