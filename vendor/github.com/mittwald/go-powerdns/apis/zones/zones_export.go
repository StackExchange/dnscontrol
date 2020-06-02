package zones

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mittwald/go-powerdns/pdnshttp"
	"net/http"
	"net/url"
)

func (c *client) ExportZone(ctx context.Context, serverID, zoneID string) ([]byte, error) {
	output := bytes.Buffer{}
	path := fmt.Sprintf("/api/v1/servers/%s/zones/%s/export", url.PathEscape(serverID), url.PathEscape(zoneID))

	err := c.httpClient.Get(ctx, path, &output)
	if err != nil {
		if e, ok := err.(pdnshttp.ErrUnexpectedStatus); ok {
			if e.StatusCode == http.StatusUnprocessableEntity {
				return nil, pdnshttp.ErrNotFound{}
			}
		}

		return nil, err
	}

	return output.Bytes(), nil
}
