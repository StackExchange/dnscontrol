package none

import "github.com/StackExchange/dnscontrol/v4/models"

// This file completes the "Registrar" interface.

// GetNameservers returns the current nameservers for a domain.
func (n None) GetNameservers(string) ([]*models.Nameserver, error) {
	return nil, nil
}

// GetRegistrarCorrections returns corrections to update registrars.
func (n None) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	return nil, nil
}
