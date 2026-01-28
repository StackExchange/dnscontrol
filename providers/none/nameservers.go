package none

import "github.com/StackExchange/dnscontrol/v4/models"

// GetNameservers returns the current nameservers for a domain.
func (n none) GetNameservers(string) ([]*models.Nameserver, error) {
	return nil, nil
}

// GetRegistrarCorrections returns corrections to update registrars.
func (n none) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	return nil, nil
}
