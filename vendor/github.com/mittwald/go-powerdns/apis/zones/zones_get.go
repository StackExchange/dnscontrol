package zones

import (
	"context"
	"fmt"
	"github.com/mittwald/go-powerdns/pdnshttp"
	"net/http"
	"net/url"
)

func (c *client) GetZone(ctx context.Context, serverID, zoneID string) (*Zone, error) {
	zone := Zone{}
	path := fmt.Sprintf("/api/v1/servers/%s/zones/%s", url.PathEscape(serverID), url.PathEscape(zoneID))

	err := c.httpClient.Get(ctx, path, &zone)
	if err != nil {
		if e, ok := err.(pdnshttp.ErrUnexpectedStatus); ok {
			if e.StatusCode == http.StatusUnprocessableEntity {
				return nil, pdnshttp.ErrNotFound{}
			}
		}

		return nil, err
	}

	return &zone, nil
}
