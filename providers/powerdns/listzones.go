package powerdns

import (
	"context"
	"strings"
)

// ListZones returns all the zones in an account
func (api *powerdnsProvider) ListZones() ([]string, error) {
	var result []string
	myZones, err := api.client.Zones().ListZones(context.Background(), api.ServerName)
	if err != nil {
		return result, err
	}
	for _, zone := range myZones {
		result = append(result, strings.TrimSuffix(zone.Name, "."))
	}
	return result, nil
}
