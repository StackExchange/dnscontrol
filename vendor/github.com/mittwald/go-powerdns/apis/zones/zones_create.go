package zones

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mittwald/go-powerdns/pdnshttp"
)

func (c *client) CreateZone(ctx context.Context, serverID string, zone Zone) (*Zone, error) {
	created := Zone{}
	path := fmt.Sprintf("/api/v1/servers/%s/zones", url.PathEscape(serverID))

	zone.ID = ""
	zone.Type = ZoneTypeZone

	if zone.Kind == 0 {
		zone.Kind = ZoneKindNative
	}

	err := c.httpClient.Post(ctx, path, &created, pdnshttp.WithJSONRequestBody(&zone))
	if err != nil {
		return nil, err
	}

	return &created, nil
}
