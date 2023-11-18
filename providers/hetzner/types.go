package hetzner

import (
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
)

type bulkCreateRecordsRequest struct {
	Records []record `json:"records"`
}

type bulkUpdateRecordsRequest struct {
	Records []record `json:"records"`
}

type createZoneRequest struct {
	Name string `json:"name"`
}

type getAllRecordsResponse struct {
	Records []record `json:"records"`
	Meta    struct {
		Pagination struct {
			LastPage     int `json:"last_page"`
			TotalEntries int `json:"total_entries"`
		} `json:"pagination"`
	} `json:"meta"`
}

type getAllZonesResponse struct {
	Zones []zone `json:"zones"`
	Meta  struct {
		Pagination struct {
			LastPage     int `json:"last_page"`
			TotalEntries int `json:"total_entries"`
		} `json:"pagination"`
	} `json:"meta"`
}

type record struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	TTL    *uint32 `json:"ttl"`
	Type   string  `json:"type"`
	Value  string  `json:"value"`
	ZoneID string  `json:"zone_id"`
}

type zone struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	NameServers []string `json:"ns"`
	TTL         uint32   `json:"ttl"`
}

func fromRecordConfig(in *models.RecordConfig, zone *zone) record {
	r := record{
		Name:   in.GetLabel(),
		Type:   in.Type,
		Value:  in.GetTargetCombined(),
		TTL:    &in.TTL,
		ZoneID: zone.ID,
	}

	if r.Type == "TXT" && (in.GetTargetTXTSegmentCount() == 1) {
		// HACK: HETZNER rejects values that fit into 255 bytes w/o quotes,
		//  but do not fit w/ added quotes (via GetTargetCombined()).
		// Sending the raw, non-quoted value works for the comprehensive
		//  suite of integrations tests.
		// The HETZNER validation does not provide helpful error messages.
		// {"error":{"message":"422 Unprocessable Entity: missing: ; ","code":422}}
		// Last checked: 2023-04-01
		valueNotQuoted := in.GetTargetTXTSegmented()[0]
		if len(valueNotQuoted) == 254 || len(valueNotQuoted) == 255 {
			r.Value = valueNotQuoted
		}
	}

	return r
}

func toRecordConfig(domain string, r *record) (*models.RecordConfig, error) {
	rc := models.RecordConfig{
		Type:     r.Type,
		TTL:      *r.TTL,
		Original: r,
	}
	rc.SetLabel(r.Name, domain)

	value := r.Value
	// HACK: Hetzner is inserting a trailing space after multiple, quoted values.
	// NOTE: The actual DNS answer does not contain the space.
	// Last checked: 2023-04-01
	if r.Type == "TXT" && len(value) > 0 && value[len(value)-1] == ' ' {
		// Per RFC 1035 spaces outside quoted values are irrelevant.
		value = strings.TrimRight(value, " ")
	}
	return &rc, rc.PopulateFromString(r.Type, value, domain)
}
