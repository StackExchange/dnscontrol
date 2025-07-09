package msdns

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
)

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (client *msdnsProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, foundRecords models.Records) ([]*models.Correction, int, error) {
	var corrections []*models.Correction

	models.PostProcessRecords(foundRecords)

	changes, actualChangeCount, err := diff2.ByRecord(foundRecords, dc, nil)
	if err != nil {
		return nil, 0, err
	}

	var corr *models.Correction
	for _, change := range changes {
		msgsJoined := change.MsgsJoined
		switch change.Type {
		case diff2.REPORT:
			corr = &models.Correction{Msg: msgsJoined}
		case diff2.CREATE:
			newrec := change.New[0]
			corr = &models.Correction{
				Msg: msgsJoined,
				F: func() error {
					return client.createOneRecord(client.dnsserver, dc.Name, newrec)
				},
			}
		case diff2.CHANGE:
			oldrec := change.Old[0]
			newrec := change.New[0]
			var f func(dnsserver string, zonename string, oldrec *models.RecordConfig, newrec *models.RecordConfig) error
			if change.HintOnlyTTL && change.HintRecordSetLen1 {
				// If we're only changing the TTL, and there is exactly one
				// record of type oldrec.Type at this label, then we can do the
				// TTL change in one command instead of deleting and re-creating
				// the record.
				f = client.modifyRecordTTL
			} else {
				f = client.modifyOneRecord
			}
			corr = &models.Correction{
				Msg: msgsJoined,
				F: func() error {
					return f(client.dnsserver, dc.Name, oldrec, newrec)
				},
			}
		case diff2.DELETE:
			oldrec := change.Old[0]
			corr = &models.Correction{
				Msg: msgsJoined,
				F: func() error {
					return client.deleteOneRecord(client.dnsserver, dc.Name, oldrec)
				},
			}
		default:
			panic(fmt.Sprintf("unhandled change.Type %s", change.Type))
		}

		corrections = append(corrections, corr)
	}

	return corrections, actualChangeCount, nil
}

func (client *msdnsProvider) deleteOneRecord(dnsserver, zonename string, oldrec *models.RecordConfig) error {
	return client.shell.RecordDelete(dnsserver, zonename, oldrec)
}

func (client *msdnsProvider) createOneRecord(dnsserver, zonename string, newrec *models.RecordConfig) error {
	return client.shell.RecordCreate(dnsserver, zonename, newrec)
}

func (client *msdnsProvider) modifyOneRecord(dnsserver, zonename string, oldrec, newrec *models.RecordConfig) error {
	return client.shell.RecordModify(dnsserver, zonename, oldrec, newrec)
}

func (client *msdnsProvider) modifyRecordTTL(dnsserver, zonename string, oldrec, newrec *models.RecordConfig) error {
	return client.shell.RecordModifyTTL(dnsserver, zonename, oldrec, newrec.TTL)
}
