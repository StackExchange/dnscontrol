package gandi

import (
	"github.com/tiramiseb/go-gandi-livedns/gandi_domain"
	"github.com/tiramiseb/go-gandi-livedns/gandi_livedns"
)

type Config struct {
	SharingID string
	Debug     bool
}

func NewDomainClient(apikey string, config Config) *gandi_domain.Domain {
	return gandi_domain.New(apikey, config.SharingID, config.Debug)
}

func NewLiveDNSClient(apikey string, config Config) *gandi_livedns.LiveDNS {
	return gandi_livedns.New(apikey, config.SharingID, config.Debug)
}
