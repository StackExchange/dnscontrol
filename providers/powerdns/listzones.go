package powerdns

import (
	"github.com/StackExchange/dnscontrol/v3/internal/dnscontrol"
	"strings"
)

// ListZones returns all the zones in an account
func (dsp *powerdnsProvider) ListZones(ctx dnscontrol.Context) ([]string, error) {
	var result []string
	myZones, err := dsp.client.Zones().ListZones(ctx, dsp.ServerName)
	if err != nil {
		return result, err
	}
	for _, zone := range myZones {
		result = append(result, strings.TrimSuffix(zone.Name, "."))
	}
	return result, nil
}
