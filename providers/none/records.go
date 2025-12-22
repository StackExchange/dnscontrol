package none

import "github.com/StackExchange/dnscontrol/v4/models"

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (n none) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	return nil, nil
}

// GetZoneRecordsCorrections gets the records of a zone and returns them in RecordConfig format.
func (n none) GetZoneRecordsCorrections(dc *models.DomainConfig, records models.Records) ([]*models.Correction, int, error) {
	return nil, 0, nil
}

// // GetDomainCorrections returns corrections to update a domain.
// func (n none) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
// 	return nil, nil
// }
