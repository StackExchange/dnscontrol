package domainnameshop

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
)

func (api *domainNameShopProvider) GetZoneRecords(domain string) (models.Records, error) {
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

func (api *domainNameShopProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()
	existingRecords, err := api.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	// Normalize
	models.PostProcessRecords(existingRecords)

	// Merge TXT strings to one string
	for _, rc := range dc.Records {
		if rc.HasFormatIdenticalToTXT() {
			rc.SetTargetTXT(strings.Join(rc.TxtStrings, ""))
		}
	}

	// Domainnameshop doesn't allow arbitrary TTLs they must be a multiple of 60.
	for _, record := range dc.Records {
		record.TTL = fixTTL(record.TTL)
	}

	differ := diff.New(dc)
	_, create, delete, modify, err := differ.IncrementalDiff(existingRecords)
	if err != nil {
		return nil, err
	}

	var corrections = []*models.Correction{}

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
		domainName := strings.Replace(r.Desired.GetLabelFQDN(), r.Desired.GetLabel()+".", "", -1)
		dnsR, err := api.fromRecordConfig(domainName, r.Desired)
		if err != nil {
			return nil, err
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
			return nil, err
		}

		dnsR.ID = r.Existing.Original.(*domainNameShopRecord).ID

		corr := &models.Correction{
			Msg: r.String(),
			F:   func() error { return api.UpdateRecord(dnsR) },
		}

		corrections = append(corrections, corr)
	}

	return corrections, nil
}

const minAllowedTTL = 60
const maxAllowedTTL = 604800
const multiplierTTL = 60

func (api *domainNameShopProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	ns, err := api.getNS(domain)
	if err != nil {
		return nil, err
	}
	return models.ToNameservers(ns)
}

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
