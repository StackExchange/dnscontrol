package hetzner

import (
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/decode"
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
		Value:  in.GetTargetCombined(),
		TTL:    &ttl,
		ZoneID: zone.ID,
	}

	if in.HasFormatIdenticalToTXT() && len(in.TxtStrings) == 1 {
		// HACK: HETZNER rejects values that fit into 255 bytes w/o quotes,
		//  but do not fit w/ added quotes (via GetTargetCombined()).
		// Sending the raw, non-quoted value works for the comprehensive
		//  suite of integrations tests.
		// The HETZNER validation does not provide helpful error messages.
		// {"error":{"message":"422 Unprocessable Entity: missing: ; ","code":422}}
		valueNotQuoted := in.TxtStrings[0]
		if len(valueNotQuoted) == 254 || len(valueNotQuoted) == 255 {
			record.Value = valueNotQuoted
		}
	}

	return record
}

func toRecordConfig(domain string, record *record) (*models.RecordConfig, error) {
	rc := &models.RecordConfig{
		Type:     record.Type,
		TTL:      uint32(*record.TTL),
		Original: record,
	}
	rc.SetLabel(record.Name, domain)

	if !rc.HasFormatIdenticalToTXT() {
		return rc, rc.PopulateFromString(record.Type, record.Value, domain)
	}

	value := record.Value
	// HACK: Hetzner is inserting a trailing space after multiple, quoted values.
	// NOTE: The actual DNS answer does not contain the space.
	// Per RFC 1035 spaces outside quoted values are irrelevant.
	value = strings.TrimRight(value, " ")

	if !decode.IsQuoted(value) {
		// This is a simple value that was set via some other client/GUI; Or
		//  this is a 254/255 long string -- see the HACK in encoding section.
		return rc, rc.SetTargetTXTs([]string{value})
	}
	s, err := decode.QuotedFields(value)
	if err != nil {
		return nil, err
	}
	return rc, rc.SetTargetTXTs(s)
}
