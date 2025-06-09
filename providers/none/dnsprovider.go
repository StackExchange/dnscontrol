package none

import (
	"github.com/StackExchange/dnscontrol/v4/models"
)

// GetZoneRecordsCorrections gets the records of a zone and returns them in RecordConfig format.
func (n None) GetZoneRecordsCorrections(dc *models.DomainConfig, records models.Records) ([]*models.Correction, int, error) {

	/*
		// if your provider does updates one record at a time...
		changes, err := diff2.ByRecord(existing, dc, nil)

		// if your provider does updates on all the records at a set (label+type) at once...
		changes, err := diff2.ByRecordSet(existing, dc, nil)

		// if your provider does updates on all the records at a label at once...
		changes, err := diff2.ByLabel(existing, dc, nil)

		if err != nil {
			return nil, err
		}

		var corrections []*models.Correction

		for _, change := range changes {
			switch change.Type {
			case diff2.REPORT:
				corr = change.CreateMessage()
			case diff2.CREATE:
				corr = change.CreateCorrection(func() error { return c.createRecord(FILL_IN) })
			case diff2.CHANGE:
				corr = change.CreateCorrection(func() error { return c.modifyRecord(FILL_IN) })
			case diff2.DELETE:
				corr = change.CreateCorrection(func() error { return c.deleteRecord(FILL_IN) })
			default:
				panic("unhandled change.TYPE %s", change.Type)
			}

			corrections = append(corrections, corr)
		}

		return corrections, nil
	*/

	return nil, 0, nil
}

//// GetDomainCorrections returns corrections to update a domain.
//func (n None) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
//return nil, nil
//}
