package zones

import (
	"context"
	"fmt"
	"net/url"
)

func (c *client) NotifySlaves(ctx context.Context, serverID string, zoneID string) error {
	path := fmt.Sprintf("/api/v1/servers/%s/zones/%s/notify", url.PathEscape(serverID), url.PathEscape(zoneID))

	return c.httpClient.Put(ctx, path, nil)
}
