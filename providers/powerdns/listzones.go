package powerdns

import (
	"context"
	"strings"
)

// ListZones returns all the zones in an account
func (dsp *powerdnsProvider) ListZones() ([]string, error) {
	var result []string
	myZones, err := dsp.client.Zones().ListZones(context.Background(), dsp.ServerName)
	if err != nil {
		return result, err
	}
	for _, zone := range myZones {
		result = append(result, strings.TrimSuffix(zone.Name, "."))
	}
	return result, nil
}
