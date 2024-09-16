package domainnameshop

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff"
)

func (api *domainNameShopProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	records, err := api.getDNS(domain)
	if err != nil {
		return nil, err
	}

	var existingRecords []*models.RecordConfig
	for i := range records {
		rC := toRecordConfig(domain, &records[i])
		existingRecords = append(existingRecords, rC)
	}

	return existingRecords, nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (api *domainNameShopProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, int, error) {

	// Merge TXT strings to one string
	for _, rc := range dc.Records {
		if rc.HasFormatIdenticalToTXT() {
			rc.SetTargetTXT(rc.GetTargetTXTJoined())
		}
	}

	// Domainnameshop doesn't allow arbitrary TTLs they must be a multiple of 60.
	for _, record := range dc.Records {
		record.TTL = fixTTL(record.TTL)
	}

	toReport, create, delete, modify, actualChangeCount, err := diff.NewCompat(dc).IncrementalDiff(existingRecords)
	if err != nil {
		return nil, 0, err
	}
	// Start corrections with the reports
	corrections := diff.GenerateMessageCorrections(toReport)

	// Delete record
	for _, r := range delete {
		domainID := r.Existing.Original.(*domainNameShopRecord).DomainID
		recordID := strconv.Itoa(r.Existing.Original.(*domainNameShopRecord).ID)

		corr := &models.Correction{
			Msg: fmt.Sprintf("%s, record id: %s", r.String(), recordID),
			F:   func() error { return api.deleteRecord(domainID, recordID) },
		}
		corrections = append(corrections, corr)
	}

	// Create records
	for _, r := range create {
		// Retrieve the domain name that is targeted. I.e. example.com instead of sub.example.com
		domainName := strings.Replace(r.Desired.GetLabelFQDN(), r.Desired.GetLabel()+".", "", -1)

		dnsR, err := api.fromRecordConfig(domainName, r.Desired)
		if err != nil {
			return nil, 0, err
		}

		corr := &models.Correction{
			Msg: r.String(),
			F:   func() error { return api.CreateRecord(domainName, dnsR) },
		}

		corrections = append(corrections, corr)
	}

	for _, r := range modify {
		domainName := strings.Replace(r.Desired.GetLabelFQDN(), r.Desired.GetLabel()+".", "", -1)

		dnsR, err := api.fromRecordConfig(domainName, r.Desired)
		if err != nil {
			return nil, 0, err
		}

		dnsR.ID = r.Existing.Original.(*domainNameShopRecord).ID

		corr := &models.Correction{
			Msg: r.String(),
			F:   func() error { return api.UpdateRecord(dnsR) },
		}

		corrections = append(corrections, corr)
	}

	return corrections, actualChangeCount, nil
}

func (api *domainNameShopProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	ns, err := api.getNS(domain)
	if err != nil {
		return nil, err
	}
	return models.ToNameservers(ns)
}

const minAllowedTTL = 60
const maxAllowedTTL = 604800
const multiplierTTL = 60

func fixTTL(ttl uint32) uint32 {
	// if the TTL is larger than the largest allowed value, return the largest allowed value
	if ttl > maxAllowedTTL {
		return maxAllowedTTL
	} else if ttl < 60 {
		return minAllowedTTL
	}

	// Return closest rounded down possible

	return (ttl / multiplierTTL) * multiplierTTL
}
