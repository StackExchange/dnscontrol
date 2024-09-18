package rwth

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff"
)

// RWTHDefaultNs is the default DNS NS for this provider.
var RWTHDefaultNs = []string{"dns-1.dfn.de", "dns-2.dfn.de", "zs1.rz.rwth-aachen.de", "zs2.rz.rwth-aachen.de"}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *rwthProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	records, err := api.getAllRecords(domain)
	if err != nil {
		return nil, err
	}
	foundRecords := models.Records{}
	for i := range records {
		foundRecords = append(foundRecords, &records[i])
	}
	return foundRecords, nil
}

// GetNameservers returns the default nameservers for RWTH.
func (api *rwthProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.ToNameservers(RWTHDefaultNs)
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (api *rwthProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, int, error) {
	domain := dc.Name

	toReport, create, del, modify, actualChangeCount, err := diff.NewCompat(dc).IncrementalDiff(existingRecords)
	if err != nil {
		return nil, 0, err
	}
	// Start corrections with the reports
	corrections := diff.GenerateMessageCorrections(toReport)

	for _, d := range create {
		des := d.Desired
		corrections = append(corrections, &models.Correction{
			Msg: d.String(),
			F:   func() error { return api.createRecord(des) },
		})
	}
	for _, d := range del {
		existingRecord := d.Existing.Original.(RecordReply)
		corrections = append(corrections, &models.Correction{
			Msg: d.String(),
			F:   func() error { return api.destroyRecord(existingRecord) },
		})
	}
	for _, d := range modify {
		rec := d.Desired
		existingID := d.Existing.Original.(RecordReply).ID
		corrections = append(corrections, &models.Correction{
			Msg: d.String(),
			F:   func() error { return api.updateRecord(existingID, *rec) },
		})
	}

	// And deploy if any corrections were applied
	if len(corrections) > 0 {
		corrections = append(corrections, &models.Correction{
			Msg: fmt.Sprintf("Deploy zone %s", domain),
			F:   func() error { return api.deployZone(domain) },
		})
	}

	return corrections, actualChangeCount, nil
}
