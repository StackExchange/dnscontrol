package msdns

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v3/pkg/txtutil"
)

// GetZoneRecordsCorrections gets existing records, diffs them against existing, and returns corrections.
func (client *msdnsProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, foundRecords models.Records) ([]*models.Correction, error) {

	// Normalize
	models.PostProcessRecords(foundRecords)
	txtutil.SplitSingleLongTxt(dc.Records) // Autosplit long TXT records

	var corrections []*models.Correction
	if !diff2.EnableDiff2 {
		differ := diff.New(dc)
		_, creates, dels, modifications, err := differ.IncrementalDiff(foundRecords)
		if err != nil {
			return nil, err
		}

		// Generate changes.
		for _, del := range dels {
			corrections = append(corrections, client.deleteRec(client.dnsserver, dc.Name, del))
		}
		for _, cre := range creates {
			corrections = append(corrections, client.createRec(client.dnsserver, dc.Name, cre)...)
		}
		for _, m := range modifications {
			corrections = append(corrections, client.modifyRec(client.dnsserver, dc.Name, m))
		}
		return corrections, nil
	}

	changes, err := diff2.ByRecord(foundRecords, dc, nil)
	if err != nil {
		return nil, err
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
			corr = &models.Correction{
				Msg: msgsJoined,
				F: func() error {
					return client.modifyOneRecord(client.dnsserver, dc.Name, oldrec, newrec)
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

	return corrections, nil
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

func (client *msdnsProvider) deleteRec(dnsserver, domainname string, cor diff.Correlation) *models.Correction {
	rec := cor.Existing
	return &models.Correction{
		Msg: cor.String(),
		F: func() error {
			return client.shell.RecordDelete(dnsserver, domainname, rec)
		},
	}
}

func (client *msdnsProvider) createRec(dnsserver, domainname string, cre diff.Correlation) []*models.Correction {
	rec := cre.Desired
	arr := []*models.Correction{{
		Msg: cre.String(),
		F: func() error {
			return client.shell.RecordCreate(dnsserver, domainname, rec)
		},
	}}
	return arr
}

func (client *msdnsProvider) modifyRec(dnsserver, domainname string, m diff.Correlation) *models.Correction {
	old, rec := m.Existing, m.Desired
	return &models.Correction{
		Msg: m.String(),
		F: func() error {
			return client.shell.RecordModify(dnsserver, domainname, old, rec)
		},
	}
}
