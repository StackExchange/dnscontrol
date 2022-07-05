package powerdns

import (
	"context"
	"github.com/StackExchange/dnscontrol/v3/internal/dnscontrol"
	"strings"
)

// ListZones returns all the zones in an account
func (dsp *powerdnsProvider) ListZones(_ dnscontrol.Context) ([]string, error) {
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
