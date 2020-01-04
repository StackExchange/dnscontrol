package livedns

import (
	"github.com/tiramiseb/go-gandi-livedns"
	"github.com/tiramiseb/go-gandi-livedns/client"
)

type LiveDNS struct {
	client client.Gandi
}

func New(apikey string, config *gandi.Config) *LiveDNS {
	client := client.New(apikey,  config)
	client.SetEndpoint("livedns/")
	return &LiveDNS{client: *client}
}

func NewFromClient(g client.Gandi) *LiveDNS {
	g.SetEndpoint("livedns/")
	return &LiveDNS{client: g}
}
