package livedns

import (
	"github.com/tiramiseb/go-gandi/internal/client"
)

type LiveDNS struct {
	client client.Gandi
}

func New(apikey string, sharingid string, debug bool) *LiveDNS {
	client := client.New(apikey, sharingid, debug)
	client.SetEndpoint("livedns/")
	return &LiveDNS{client: *client}
}

func NewFromClient(g client.Gandi) *LiveDNS {
	g.SetEndpoint("livedns/")
	return &LiveDNS{client: g}
}
