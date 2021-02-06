package msdns

import "github.com/StackExchange/dnscontrol/v3/models"

func (c *msdnsProvider) GetNameservers(string) ([]*models.Nameserver, error) {
	// TODO: If using AD for publicly hosted zones, probably pull these from config.
	return nil, nil
}
