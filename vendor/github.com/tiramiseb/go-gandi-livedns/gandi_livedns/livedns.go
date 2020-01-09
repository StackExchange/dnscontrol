package gandi_livedns

import (
	"github.com/tiramiseb/go-gandi-livedns/gandi_config"
	"github.com/tiramiseb/go-gandi-livedns/internal/client"
)

type LiveDNS struct {
	client client.Gandi
}

func New(apikey string, config gandi_config.Config) *LiveDNS {
	client := client.New(apikey, config)
	client.SetEndpoint("livedns/")
	return &LiveDNS{client: *client}
}

func NewFromClient(g client.Gandi) *LiveDNS {
	g.SetEndpoint("livedns/")
	return &LiveDNS{client: g}
}
