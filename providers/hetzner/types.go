package hetzner

import (
	"github.com/StackExchange/dnscontrol/v3/models"
)

type createRecordRequest struct {
	Name   string `json:"name"`
	TTL    int    `json:"ttl"`
	Type   string `json:"type"`
	Value  string `json:"value"`
	ZoneId string `json:"zone_id"`
}

type createZoneRequest struct {
	Name string `json:"name"`
}

type getAllRecordsResponse struct {
	Records []Record `json:"records"`
	Meta    struct {
		Pagination struct {
			LastPage int `json:"last_page"`
		} `json:"pagination"`
	} `json:"meta"`
}

type getAllZonesResponse struct {
	Zones []Zone `json:"zones"`
	Meta  struct {
		Pagination struct {
			LastPage int `json:"last_page"`
		} `json:"pagination"`
	} `json:"meta"`
}

type Record struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	TTL    *int   `json:"ttl"`
	Type   string `json:"type"`
	Value  string `json:"value"`
	ZoneId string `json:"zone_id"`
}

type Zone struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	NameServers []string `json:"ns"`
	TTL         int      `json:"ttl"`
}

func fromRecordConfig(in *models.RecordConfig, zone Zone) *Record {
	ttl := int(in.TTL)
	record := &Record{
		Name:   in.GetLabel(),
		Type:   in.Type,
		Value:  in.GetTargetField(),
		TTL:    &ttl,
		ZoneId: zone.Id,
	}

	switch record.Type {
	case "TXT":
		// Cannot use `in.GetTargetCombined()` for TXTs:
		// Their validation would complain about a missing `;`.
		// Test case: single_TXT:Create_a_255-byte_TXT
		// {"error":{"message":"422 Unprocessable Entity: missing: ; ","code":422}}
		record.Value = in.GetTargetField()
	default:
		record.Value = in.GetTargetCombined()
	}

	return record
}

func toRecordConfig(domain string, record *Record) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type:     record.Type,
		TTL:      uint32(*record.TTL),
		Original: record,
	}
	rc.SetLabel(record.Name, domain)

	_ = rc.PopulateFromString(record.Type, record.Value, domain)

	return rc
}
