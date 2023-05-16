package zonerecs

import (
	"github.com/StackExchange/dnscontrol/v3/models"
)

// CorrectZoneRecords calls both GetZoneRecords, does any
// post-processing, and then calls GetZoneRecordsCorrections.  The
// name sucks because all the good names were taken.
func CorrectZoneRecords(driver models.DNSProvider, dc *models.DomainConfig) ([]*models.Correction, error) {

	existingRecords, err := driver.GetZoneRecords(dc.Name, dc.Metadata)
	if err != nil {
		return nil, err
	}

	// downcase
	models.Downcase(existingRecords)

	// Copy dc so that any corrections code that wants to
	// modify the records may. For example, if the provider only
	// supports certain TTL values, it will adjust the ones in
	// dc.Records.
	dc, err = dc.Copy()
	if err != nil {
		return nil, err
	}

	// punycode
	dc.Punycode()
	// FIXME(tlim) It is a waste to PunyCode every iteration.
	// This should be moved to where the JavaScript is processed.

	everything, err := driver.GetZoneRecordsCorrections(dc, existingRecords)
	reports, corrections := splitReportsAndCorrections(everything)
	return rerports, corrections, err
}

func splitReportsAndCorrections(everything []*models.Correction) (reports, corrections []*models.Correction) {
	for i := range everything {
		if everything[i].F == nil {
			reports = append(reports, everything[i])
		} else {
			corrections = append(corrections, everything[i])
		}
	}
	return reports, corrections
}
