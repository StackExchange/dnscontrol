package msdns

import "github.com/StackExchange/dnscontrol/v4/models"

func (client *msdnsProvider) GetNameservers(string) ([]*models.Nameserver, error) {
	// TODO: If using AD for publicly hosted zones, probably pull these from config.
	return nil, nil
}
