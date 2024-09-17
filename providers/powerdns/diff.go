package powerdns

import (
	"context"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/fatih/color"
	"github.com/mittwald/go-powerdns/apis/zones"
)

func (dsp *powerdnsProvider) getDiff2DomainCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, int, error) {
	changes, actualChangeCount, err := diff2.ByRecordSet(existing, dc, nil)
	if err != nil {
		return nil, 0, err
	}

	var corrections []*models.Correction
	var changeMsgs []string
	var rrChangeSets []zones.ResourceRecordSet
	var deleteMsgs []string
	var rrDeleteSets []zones.ResourceRecordSet

	// for pretty alignment, add an empty string
	changeMsgs = append(changeMsgs, color.YellowString("± BATCHED CHANGE/CREATEs for %s", dc.Name))
	deleteMsgs = append(deleteMsgs, color.RedString("- BATCHED DELETEs for %s", dc.Name))

	for _, change := range changes {
		labelName := canonical(change.Key.NameFQDN)
		labelType := change.Key.Type

		switch change.Type {
		case diff2.REPORT:
			corrections = append(corrections, &models.Correction{Msg: change.MsgsJoined})
		case diff2.CREATE, diff2.CHANGE:
			labelTTL := int(change.New[0].TTL)
			records := buildRecordList(change)

			rrChangeSets = append(rrChangeSets, zones.ResourceRecordSet{
				Name:    labelName,
				Type:    labelType,
				TTL:     labelTTL,
				Records: records,
				// ChangeType is not needed since zone API sets it when calling Add
			})
			changeMsgs = append(changeMsgs, change.MsgsJoined)
		case diff2.DELETE:
			rrDeleteSets = append(rrDeleteSets, zones.ResourceRecordSet{
				Name: labelName,
				Type: labelType,
				// ChangeType is not needed since zone API sets it when calling Remove
			})
			deleteMsgs = append(deleteMsgs, change.MsgsJoined)
		default:
			panic(fmt.Sprintf("unhandled change.Type %s", change.Type))
		}
	}

	// only append a Correction if there are any, otherwise causes an error when sending an empty rrset
	if len(rrDeleteSets) > 0 {
		corrections = append(corrections, &models.Correction{
			Msg: strings.Join(deleteMsgs, "\n"),
			F: func() error {
				return dsp.client.Zones().RemoveRecordSetsFromZone(context.Background(), dsp.ServerName, canonical(dc.Name), rrDeleteSets)
			},
		})
	}
	if len(rrChangeSets) > 0 {
		corrections = append(corrections, &models.Correction{
			Msg: strings.Join(changeMsgs, "\n"),
			F: func() error {
				return dsp.client.Zones().AddRecordSetsToZone(context.Background(), dsp.ServerName, canonical(dc.Name), rrChangeSets)
			},
		})
	}
	return corrections, actualChangeCount, nil
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
