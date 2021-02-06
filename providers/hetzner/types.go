package hetzner

import (
	"github.com/StackExchange/dnscontrol/v3/models"
)

type bulkCreateRecordsRequest struct {
	Records []record `json:"records"`
}

type bulkUpdateRecordsRequest struct {
	Records []record `json:"records"`
}

type createRecordRequest struct {
	Name   string `json:"name"`
	TTL    int    `json:"ttl"`
	Type   string `json:"type"`
	Value  string `json:"value"`
	ZoneID string `json:"zone_id"`
}

type createZoneRequest struct {
	Name string `json:"name"`
}

type getAllRecordsResponse struct {
	Records []record `json:"records"`
	Meta    struct {
		Pagination struct {
			LastPage int `json:"last_page"`
		} `json:"pagination"`
	} `json:"meta"`
}

type getAllZonesResponse struct {
	Zones []zone `json:"zones"`
	Meta  struct {
		Pagination struct {
			LastPage int `json:"last_page"`
		} `json:"pagination"`
	} `json:"meta"`
}

type record struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	TTL    *int   `json:"ttl"`
	Type   string `json:"type"`
	Value  string `json:"value"`
	ZoneID string `json:"zone_id"`
}

type zone struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	NameServers []string `json:"ns"`
	TTL         int      `json:"ttl"`
}

func fromRecordConfig(in *models.RecordConfig, zone *zone) *record {
	ttl := int(in.TTL)
	record := &record{
		Name:   in.GetLabel(),
		Type:   in.Type,
		Value:  in.GetTargetField(),
		TTL:    &ttl,
		ZoneID: zone.ID,
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

func toRecordConfig(domain string, record *record) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type:     record.Type,
		TTL:      uint32(*record.TTL),
		Original: record,
	}
	rc.SetLabel(record.Name, domain)

	_ = rc.PopulateFromString(record.Type, record.Value, domain)

	return rc
}
