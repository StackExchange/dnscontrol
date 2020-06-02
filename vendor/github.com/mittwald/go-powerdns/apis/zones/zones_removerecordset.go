package zones

import (
	"context"
	"fmt"
	"github.com/mittwald/go-powerdns/pdnshttp"
	"net/url"
)

func (c *client) RemoveRecordSetFromZone(ctx context.Context, serverID string, zoneID string, name string, recordType string) error {
	path := fmt.Sprintf("/api/v1/servers/%s/zones/%s", url.PathEscape(serverID), url.PathEscape(zoneID))

	set := ResourceRecordSet{
		Name:       name,
		Type:       recordType,
		ChangeType: ChangeTypeDelete,
	}

	patch := Zone{
		ResourceRecordSets: []ResourceRecordSet{set},
	}

	return c.httpClient.Patch(ctx, path, nil, pdnshttp.WithJSONRequestBody(&patch))
}
