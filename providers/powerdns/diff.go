package powerdns

import (
	"context"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff2"
	"github.com/mittwald/go-powerdns/apis/zones"
)

func (dsp *powerdnsProvider) getDiff1DomainCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, error) {
	// create record diff by group
	keysToUpdate, err := (diff.New(dc)).ChangedGroups(existing)
	if err != nil {
		return nil, err
	}
	desiredRecords := dc.Records.GroupedByKey()

	var cuCorrections []*models.Correction
	var dCorrections []*models.Correction

	// add create/update and delete corrections separately
	for label, msgs := range keysToUpdate {
		labelName := canonical(label.NameFQDN)
		labelType := label.Type
		msgJoined := strings.Join(msgs, "\n   ")

		if _, ok := desiredRecords[label]; !ok {
			// no record found so delete it
			dCorrections = append(dCorrections, &models.Correction{
				Msg: msgJoined,
				F: func() error {
					return dsp.client.Zones().RemoveRecordSetFromZone(context.Background(), dsp.ServerName, dc.Name, labelName, labelType)
				},
			})
		} else {
			// record found so create or update it
			ttl := desiredRecords[label][0].TTL
			var records []zones.Record
			for _, recordContent := range desiredRecords[label] {
				records = append(records, zones.Record{
					Content: recordContent.GetTargetCombined(),
				})
			}
			cuCorrections = append(cuCorrections, &models.Correction{
				Msg: msgJoined,
				F: func() error {
					return dsp.client.Zones().AddRecordSetToZone(context.Background(), dsp.ServerName, dc.Name, zones.ResourceRecordSet{
						Name:       labelName,
						Type:       labelType,
						TTL:        int(ttl),
						Records:    records,
						ChangeType: zones.ChangeTypeReplace,
					})
				},
			})
		}
	}

	// append corrections in the right order
	// delete corrections must be run first to avoid correlations with existing RR
	var corrections []*models.Correction
	corrections = append(corrections, dCorrections...)
	corrections = append(corrections, cuCorrections...)

	return corrections, nil
}

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
		case diff2.CREATE, diff2.CHANGE, diff2.MODIFYTTL:
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
