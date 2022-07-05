package msdns

import (
	"github.com/StackExchange/dnscontrol/v3/internal/dnscontrol"
	"github.com/StackExchange/dnscontrol/v3/models"
)

func (client *msdnsProvider) GetNameservers(dnscontrol.Context, string) ([]*models.Nameserver, error) {
	// TODO: If using AD for publicly hosted zones, probably pull these from config.
	return nil, nil
}
