package zones

import (
	"context"
	"fmt"
	"github.com/mittwald/go-powerdns/pdnshttp"
	"net/url"
)

func (c *client) AddRecordSetToZone(ctx context.Context, serverID string, zoneID string, set ResourceRecordSet) error {
	path := fmt.Sprintf("/api/v1/servers/%s/zones/%s", url.PathEscape(serverID), url.PathEscape(zoneID))

	set.ChangeType = ChangeTypeReplace
	patch := Zone{
		ResourceRecordSets: []ResourceRecordSet{
			set,
		},
	}

	return c.httpClient.Patch(ctx, path, nil, pdnshttp.WithJSONRequestBody(&patch))
}
