package alidns

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"golang.org/x/net/idna"
)

// nativeToRecord converts an Alibaba Cloud DNS record to a RecordConfig.
func nativeToRecord(r *alidns.Record, domain string) (*models.RecordConfig, error) {
	rc := &models.RecordConfig{
		TTL:      uint32(r.TTL),
		Original: r,
	}
	label, err := idna.ToASCII(r.RR)
	if err != nil {
		return nil, fmt.Errorf("failed to convert label to ASCII: %w", err)
	}
	rc.SetLabel(label, domain)

	// Normalize CNAME, MX, NS records with trailing dot to be consistent with FQDN format.
	value := r.Value
	if r.Type == "CNAME" || r.Type == "MX" || r.Type == "NS" || r.Type == "SRV" {
		if value != "" && value != "." && !strings.HasSuffix(value, ".") {
			value = value + "."
		}
	}

	switch r.Type {
	case "MX":
		if err := rc.SetTargetMX(uint16(r.Priority), value); err != nil {
			return nil, fmt.Errorf("unparsable MX record received from ALIDNS: %w", err)
		}
	case "SRV":
		// SRV records in Alibaba Cloud: Priority and Weight are in separate fields,
		// Value contains "port target" (e.g., "5060 sipserver.example.com")
		if err := rc.PopulateFromString(r.Type, fmt.Sprintf("%d %d %s", r.Priority, r.Weight, r.Value), domain); err != nil {
			return nil, fmt.Errorf("unparsable SRV record received from ALIDNS: %w", err)
		}
	case "CAA":
		// CAA format in Alibaba: "0 issue letsencrypt.org"
		if err := rc.PopulateFromString(r.Type, r.Value, domain); err != nil {
			return nil, fmt.Errorf("unparsable CAA record received from ALIDNS: %w", err)
		}
	case "TXT":
		if err := rc.SetTargetTXT(r.Value); err != nil {
			return nil, fmt.Errorf("unparsable TXT record received from ALIDNS: %w", err)
		}
	default:
		rc.Type = r.Type
		if err := rc.SetTarget(value); err != nil {
			return nil, fmt.Errorf("unparsable record received from ALIDNS: %w", err)
		}
	}

	return rc, nil
}

// recordToNativeContent converts a RecordConfig to the Value format expected by Alibaba Cloud DNS API.
func recordToNativeContent(r *models.RecordConfig) string {
	switch r.Type {
	case "MX":
		return r.GetTargetField()
	case "SRV":
		// Alibaba Cloud SRV format: "weight port target"
		return fmt.Sprintf("%d %d %s", r.SrvWeight, r.SrvPort, r.GetTargetField())
	case "CAA":
		return fmt.Sprintf("%d %s %s", r.CaaFlag, r.CaaTag, r.GetTargetField())
	case "TXT":
		return r.GetTargetTXTJoined()
	default:
		return r.GetTargetField()
	}
}

// recordToNativePriority returns the priority value for MX and SRV records.
func recordToNativePriority(r *models.RecordConfig) int64 {
	switch r.Type {
	case "MX":
		return int64(r.MxPreference)
	case "SRV":
		return int64(r.SrvPriority)
	default:
		return 0
	}
}
