package zones

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mittwald/go-powerdns/pdnshttp"
)

func (c *client) ListZones(ctx context.Context, serverID string) ([]Zone, error) {
	zones := make([]Zone, 0)
	path := fmt.Sprintf("/api/v1/servers/%s/zones", url.PathEscape(serverID))

	err := c.httpClient.Get(ctx, path, &zones)
	if err != nil {
		return nil, err
	}

	return zones, nil
}

func (c *client) ListZone(ctx context.Context, serverID string, zoneID string) ([]Zone, error) {
	zones := make([]Zone, 0)
	path := fmt.Sprintf("/api/v1/servers/%s/zones", url.PathEscape(serverID))

	err := c.httpClient.Get(ctx, path, &zones, pdnshttp.WithQueryValue("zone", zoneID))
	if err != nil {
		return nil, err
	}

	return zones, nil
}
