package gandi

import (
	"github.com/tiramiseb/go-gandi/domain"
	"github.com/tiramiseb/go-gandi/livedns"
)

type Config struct {
	SharingID string
	Debug     bool
}

func NewDomainClient(apikey string, config Config) *domain.Domain {
	return domain.New(apikey, config.SharingID, config.Debug)
}

func NewLiveDNSClient(apikey string, config Config) *livedns.LiveDNS {
	return livedns.New(apikey, config.SharingID, config.Debug)
}
