package huaweicloud

import (
	"fmt"
	"net/url"
	"slices"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
)

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *huaweicloudProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	if err := c.getZones(); err != nil {
		return nil, err
	}
	zoneID, ok := c.zoneIDByDomain[domain]
	if !ok {
		return nil, fmt.Errorf("zone %s not found", domain)
	}
	records, err := c.fetchZoneRecordsFromRemote(zoneID)
	if err != nil {
		return nil, err
	}

	// Convert rrsets to DNSControl's RecordConfig
	existingRecords := []*models.RecordConfig{}
	for _, rec := range *records {
		if *rec.Type == "SOA" {
			continue
		}
		nativeRecords, err := nativeToRecords(&rec, domain)
		if err != nil {
			return nil, err
		}
		existingRecords = append(existingRecords, nativeRecords...)
	}

	return existingRecords, nil
}

// GenerateDomainCorrections takes the desired and existing records
// and produces a Correction list.  The correction list is simply
// a list of functions to call to actually make the desired
// correction, and a message to output to the user when the change is
// made.
func (c *huaweicloudProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, error) {
	if err := c.getZones(); err != nil {
		return nil, err
	}
	zoneID, ok := c.zoneIDByDomain[dc.Name]
	if !ok {
		return nil, fmt.Errorf("zone %s not found", dc.Name)
	}

	// Make delete happen earlier than creates & updates.
	var corrections []*models.Correction
	var deletions []*models.Correction
	var reports []*models.Correction

	changes, err := diff2.ByRecordSet(existing, dc, nil)
	if err != nil {
		return nil, err
	}

	for _, change := range changes {
		msg := change.MsgsJoined
		records := recordsToNative(change.New, change.Key)
		rrsetsID := getRRSetIDFromRecords(change.Old)

		switch change.Type {
		case diff2.REPORT:
			reports = append(reports, &models.Correction{Msg: msg})
		case diff2.CREATE:
			fallthrough
		case diff2.CHANGE:
			corrections = append(corrections, &models.Correction{
				Msg: msg,
				F: func() error {
					if len(rrsetsID) == 1 {
						return c.updateRRSet(zoneID, rrsetsID[0], records)
					} else {
						err := c.deleteRRSets(zoneID, rrsetsID)
						if err != nil {
							return err
						}
						return c.createRRSet(zoneID, records)
					}
				},
			})
		case diff2.DELETE:
			deletions = append(deletions, &models.Correction{
				Msg: msg,
				F: func() error {
					return c.deleteRRSets(zoneID, rrsetsID)
				},
			})
		default:
			panic(fmt.Sprintf("unhandled change.Type %s", change.Type))
		}
	}

	result := append(reports, deletions...)
	result = append(result, corrections...)
	return result, nil
}

func (c *huaweicloudProvider) deleteRRSets(zoneID string, rrsets []string) error {
	for _, rrset := range rrsets {
		deletePayload := &model.DeleteRecordSetRequest{
			ZoneId:      zoneID,
			RecordsetId: rrset,
		}
		var err error
		withRetry(func() error {
			_, err = c.client.DeleteRecordSet(deletePayload)
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *huaweicloudProvider) createRRSet(zoneID string, rc *model.ListRecordSets) error {
	createPayload := &model.CreateRecordSetRequest{
		ZoneId: zoneID,
		Body: &model.CreateRecordSetRequestBody{
			Name:    *rc.Name,
			Type:    *rc.Type,
			Ttl:     rc.Ttl,
			Records: *rc.Records,
		},
	}
	var err error
	withRetry(func() error {
		_, err = c.client.CreateRecordSet(createPayload)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *huaweicloudProvider) updateRRSet(zoneID, rrsetID string, rc *model.ListRecordSets) error {
	updatePayload := &model.UpdateRecordSetRequest{
		ZoneId:      zoneID,
		RecordsetId: rrsetID,
		Body: &model.UpdateRecordSetReq{
			Name:    *rc.Name,
			Type:    *rc.Type,
			Ttl:     rc.Ttl,
			Records: rc.Records,
		},
	}
	var err error
	withRetry(func() error {
		_, err = c.client.UpdateRecordSet(updatePayload)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func parseMarkerFromURL(link string) (string, error) {
	// Parse the marker params from the URL
	// Example: https://dns.myhuaweicloud.com/v2/zones?marker=abcdefg
	url, err := url.Parse(link)
	if err != nil {
		return "", err
	}
	marker := url.Query().Get("marker")
	if marker == "" {
		return "", fmt.Errorf("marker not found in URL %s", link)
	}
	return marker, nil
}

func (c *huaweicloudProvider) fetchZoneRecordsFromRemote(zoneID string) (*[]model.ListRecordSets, error) {
	var nextMarker *string
	existingRecords := []model.ListRecordSets{}
	availableStatus := []string{"ACTIVE", "PENDING_CREATE", "PENDING_UPDATE"}

	for {
		payload := model.ListRecordSetsByZoneRequest{
			ZoneId: zoneID,
			Marker: nextMarker,
		}
		var res *model.ListRecordSetsByZoneResponse
		var err error
		withRetry(func() error {
			res, err = c.client.ListRecordSetsByZone(&payload)
			return err
		})
		if err != nil {
			return nil, err
		}
		if res.Recordsets == nil {
			return &existingRecords, nil
		}
		for _, record := range *res.Recordsets {
			if record.Records == nil {
				continue
			}
			if !slices.Contains(availableStatus, *record.Status) {
				continue
			}
			existingRecords = append(existingRecords, record)
		}

		// if has next page, continue to get next page
		if res.Links.Next != nil {
			marker, err := parseMarkerFromURL(*res.Links.Next)
			if err != nil {
				return nil, err
			}
			nextMarker = &marker
		} else {
			return &existingRecords, nil
		}
	}
}
