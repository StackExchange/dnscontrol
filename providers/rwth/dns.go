package rwth

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v3/pkg/txtutil"
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
func (api *rwthProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, error) {
	txtutil.SplitSingleLongTxt(dc.Records) // Autosplit long TXT records
	domain := dc.Name

	var corrections []*models.Correction
	var differ diff.Differ
	if !diff2.EnableDiff2 {
		differ = diff.New(dc)
	} else {
		differ = diff.NewCompat(dc)
	}
	_, create, del, modify, err := differ.IncrementalDiff(existingRecords)
	if err != nil {
		return nil, err
	}

	for _, d := range create {
		des := d.Desired
		corrections = append(corrections, &models.Correction{
			Msg: d.String(),
			F:   func() error { return api.createRecord(dc.Name, des) },
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

	return corrections, nil
}
