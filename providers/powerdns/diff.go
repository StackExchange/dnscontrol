package powerdns

import (
	"context"
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/mittwald/go-powerdns/apis/zones"
)

func (dsp *powerdnsProvider) getDiff2DomainCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, error) {
	changes, err := diff2.ByRecordSet(existing, dc, nil)
	if err != nil {
		return nil, err
	}

	var corrections []*models.Correction

	for _, change := range changes {
		labelName := canonical(change.Key.NameFQDN)
		labelType := change.Key.Type

		switch change.Type {
		case diff2.REPORT:
			corrections = append(corrections, &models.Correction{Msg: change.MsgsJoined})
		case diff2.CREATE, diff2.CHANGE:
			labelTTL := int(change.New[0].TTL)
			records := buildRecordList(change)

			corrections = append(corrections, &models.Correction{
				Msg: change.MsgsJoined,
				F: func() error {
					return dsp.client.Zones().AddRecordSetToZone(context.Background(), dsp.ServerName, dc.Name, zones.ResourceRecordSet{
						Name:       labelName,
						Type:       labelType,
						TTL:        labelTTL,
						Records:    records,
						ChangeType: zones.ChangeTypeReplace,
					})
				},
			})
		case diff2.DELETE:
			corrections = append(corrections, &models.Correction{
				Msg: change.MsgsJoined,
				F: func() error {
					return dsp.client.Zones().RemoveRecordSetFromZone(context.Background(), dsp.ServerName, dc.Name, labelName, labelType)
				},
			})
		default:
			panic(fmt.Sprintf("unhandled change.Type %s", change.Type))
		}
	}

	return corrections, nil
}

// buildRecordList returns a list of records for the PowerDNS resource record set from a change
func buildRecordList(change diff2.Change) (records []zones.Record) {
	for _, recordContent := range change.New {
		records = append(records, zones.Record{
			Content: recordContent.GetTargetCombined(),
		})
	}
	return
}

func canonical(fqdn string) string {
	return fqdn + "."
}
