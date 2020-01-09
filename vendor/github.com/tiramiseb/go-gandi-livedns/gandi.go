package gandi

import (
	"github.com/tiramiseb/go-gandi-livedns/gandi_config"
	"github.com/tiramiseb/go-gandi-livedns/gandi_domain"
	"github.com/tiramiseb/go-gandi-livedns/gandi_livedns"
)

func NewDomainClient(apikey string, config gandi_config.Config) *gandi_domain.Domain {
	return gandi_domain.New(apikey, config)
}

func NewLiveDNSClient(apikey string, config gandi_config.Config) *gandi_livedns.LiveDNS {
	return gandi_livedns.New(apikey, config)
}
