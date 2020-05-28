package servers

import (
	"context"
	"fmt"
	"net/url"
)

func (c *client) GetServer(ctx context.Context, serverID string) (*Server, error) {
	server := Server{}
	err := c.httpClient.Get(ctx, fmt.Sprintf("/api/v1/servers/%s", url.PathEscape(serverID)), &server)

	if err != nil {
		return nil, err
	}

	return &server, err
}
