package none

import "github.com/StackExchange/dnscontrol/v4/models"

// GetZoneRecordsCorrections gets the records of a zone and returns them in RecordConfig format.
func (n None) GetZoneRecordsCorrections(dc *models.DomainConfig, records models.Records) ([]*models.Correction, int, error) {
	return nil, 0, nil
}

// GetDomainCorrections returns corrections to update a domain.
func (n None) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	return nil, nil
}
