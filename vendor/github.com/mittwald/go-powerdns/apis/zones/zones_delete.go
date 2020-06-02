package zones

import (
	"context"
	"fmt"
	"net/url"
)

func (c *client) DeleteZone(ctx context.Context, serverID string, zoneID string) error {
	path := fmt.Sprintf("/api/v1/servers/%s/zones/%s", url.PathEscape(serverID), url.PathEscape(zoneID))

	return c.httpClient.Delete(ctx, path, nil)
}
