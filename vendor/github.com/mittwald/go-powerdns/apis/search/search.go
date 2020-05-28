package search

import (
	"context"
	"fmt"
	"github.com/mittwald/go-powerdns/pdnshttp"
	"net/url"
)

func (c *client) Search(ctx context.Context, serverID, query string, max int, objectType ObjectType) (ResultList, error) {
	path := fmt.Sprintf("/api/v1/servers/%s/search-data", url.PathEscape(serverID))
	results := make(ResultList, 0)

	err := c.httpClient.Get(
		ctx,
		path,
		&results,
		pdnshttp.WithQueryValue("q", query),
		pdnshttp.WithQueryValue("max", fmt.Sprintf("%d", max)),
		pdnshttp.WithQueryValue("object_type", objectType.String()),
	)

	if err != nil {
		return nil, err
	}

	return results, nil
}
