package search

import "github.com/mittwald/go-powerdns/pdnshttp"

type client struct {
	httpClient *pdnshttp.Client
}

// New creates a new Search client
func New(hc *pdnshttp.Client) Client {
	return &client{
		httpClient: hc,
	}
}
