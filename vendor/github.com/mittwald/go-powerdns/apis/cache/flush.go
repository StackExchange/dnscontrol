package cache

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mittwald/go-powerdns/pdnshttp"
)

func (c *client) Flush(ctx context.Context, serverID string, name string) (*FlushResult, error) {
	cfr := FlushResult{}
	path := fmt.Sprintf("/api/v1/servers/%s/cache/flush", url.PathEscape(serverID))

	err := c.httpClient.Put(ctx, path, &cfr, pdnshttp.WithQueryValue("domain", name))
	if err != nil {
		return nil, err
	}

	return &cfr, nil
}
