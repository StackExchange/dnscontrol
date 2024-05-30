package gcore

// Convert the provider's native record description to models.RecordConfig.

import (
	"errors"
	"fmt"

	dnssdk "github.com/G-Core/gcore-dns-sdk-go"
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
)

// nativeToRecord takes a DNS record from G-Core and returns a native RecordConfig struct.
func nativeToRecords(n gcoreRRSetExtended, zoneName string) ([]*models.RecordConfig, error) {
	var rcs []*models.RecordConfig
	recName := n.Name
	recType := n.Type

	// Split G-Core's RRset into individual records
	for _, value := range n.Records {
		rc := &models.RecordConfig{
			TTL:      uint32(n.TTL),
			Original: n,
		}
		rc.SetLabelFromFQDN(recName, zoneName)
		switch recType {
		case "CAA": // G-Core API don't need quotes around CAA with whitespace
			if len(value.Content) != 3 {
				return nil, errors.New("incorrect number of fields in G-Core's CAA record")
			}

			parts := make([]string, len(value.Content))
			for i := range value.Content {
				parts[i] = fmt.Sprint(value.Content[i])
			}

			flag, tag, target := parts[0], parts[1], parts[2]
			if err := rc.SetTargetCAAStrings(flag, tag, target); err != nil {
				return nil, fmt.Errorf("unparsable record received from G-Core: %w", err)
			}

		case "TXT": // Avoid double quoting for TXT records
			if err := rc.SetTargetTXTs(convertSdkAnySliceToTxtSlice(value.Content)); err != nil {
				return nil, fmt.Errorf("unparsable record received from G-Core: %w", err)
			}

		case "SCVB": // GCore mistypes "SVCB" as "SCVB"
			if err := rc.PopulateFromString("SVCB", value.ContentToString(), zoneName); err != nil {
				return nil, fmt.Errorf("unparsable record received from G-Core: %w", err)
			}

		default: //  "A", "AAAA", "CAA", "NS", "CNAME", "MX", "PTR", "SRV"
			if err := rc.PopulateFromString(recType, value.ContentToString(), zoneName); err != nil {
				return nil, fmt.Errorf("unparsable record received from G-Core: %w", err)
			}
		}
		rcs = append(rcs, rc)
	}

	return rcs, nil
}

func recordsToNative(rcs []*models.RecordConfig, expectedKey models.RecordKey) *dnssdk.RRSet {
	// Merge DNSControl records into G-Core RRsets

	var result *dnssdk.RRSet

	for _, r := range rcs {
		label := r.GetLabel()
		if label == "@" {
			label = ""
		}
		key := r.Key()

		if key != expectedKey {
			continue
		}

		var rr dnssdk.ResourceRecord
		switch key.Type {
		case "CAA": // G-Core API don't need quotes around CAA with whitespace
			rr = dnssdk.ResourceRecord{
				Content: []interface{}{
					int64(r.CaaFlag),
					r.CaaTag,
					r.GetTargetField(),
				},
				Meta:    nil,
				Enabled: true,
			}
		case "TXT": // Avoid double quoting for TXT records
			rr = dnssdk.ResourceRecord{
				Content: convertTxtSliceToSdkAnySlice(r.GetTargetTXTJoined()),
				Meta:    nil,
				Enabled: true,
			}
		case "SVCB":
			// GCore mistypes "SVCB" as "SCVB"
			rr = dnssdk.ResourceRecord{
				Content: dnssdk.ContentFromValue("SCVB", r.GetTargetCombined()),
				Meta:    nil,
				Enabled: true,
			}
		default:
			rr = dnssdk.ResourceRecord{
				Content: dnssdk.ContentFromValue(key.Type, r.GetTargetCombined()),
				Meta:    nil,
				Enabled: true,
			}
		}

		if result == nil {
			result = &dnssdk.RRSet{
				TTL:     int(r.TTL),
				Filters: nil,
				Records: []dnssdk.ResourceRecord{rr},
			}
		} else {
			result.Records = append(result.Records, rr)

			if int(r.TTL) != result.TTL {
				printer.Warnf("All TTLs for a rrset (%v) must be the same. Using smaller of %v and %v.\n", key, r.TTL, result.TTL)
				if int(r.TTL) < result.TTL {
					result.TTL = int(r.TTL)
				}
			}
		}
	}

	return result
}

func convertTxtSliceToSdkAnySlice(record string) []any {
	result := []any{record}
	return result
}

func convertSdkAnySliceToTxtSlice(records []any) []string {
	result := []string{}
	for _, record := range records {
		result = append(result, record.(string))
	}
	return result
}
