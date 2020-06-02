package zones

import (
	"context"
	"fmt"
	"net/url"
)

func (c *client) RetrieveFromMaster(ctx context.Context, serverID string, zoneID string) error {
	path := fmt.Sprintf("/api/v1/servers/%s/zones/%s/axfr-retrieve", url.PathEscape(serverID), url.PathEscape(zoneID))

	return c.httpClient.Put(ctx, path, nil)
}
