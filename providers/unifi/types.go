package unifi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
)

// legacyDNSRecord represents a DNS record in the OLD UniFi API format (v2/api/site/{site}/static-dns).
// This API is available in UniFi Network 8.2+.
type legacyDNSRecord struct {
	ID         string `json:"_id,omitempty"`
	Enabled    bool   `json:"enabled"`
	Key        string `json:"key"`         // FQDN (e.g., "test.example.com")
	RecordType string `json:"record_type"` // A, AAAA, CNAME, MX, TXT, SRV, NS
	Value      string `json:"value"`       // Record value
	TTL        int    `json:"ttl"`         // 0 = default
	Port       int    `json:"port"`        // SRV port
	Priority   int    `json:"priority"`    // MX/SRV priority
	Weight     int    `json:"weight"`      // SRV weight
}

// legacyToRecord converts a UniFi legacy API record to a dnscontrol RecordConfig.
func legacyToRecord(domain string, r *legacyDNSRecord) (*models.RecordConfig, error) {
	rc := &models.RecordConfig{
		Type:     r.RecordType,
		Original: r,
	}

	// Set TTL (UniFi uses 0 for default, we map to 300)
	if r.TTL > 0 {
		rc.TTL = uint32(r.TTL)
	} else {
		rc.TTL = 300
	}

	// Set label from FQDN
	rc.SetLabelFromFQDN(r.Key, domain)

	var err error
	switch r.RecordType {
	case "A", "AAAA":
		err = rc.SetTarget(r.Value)

	case "CNAME", "NS":
		target := r.Value
		if !strings.HasSuffix(target, ".") {
			target = target + "."
		}
		err = rc.SetTarget(target)

	case "MX":
		rc.MxPreference = uint16(r.Priority)
		target := r.Value
		if !strings.HasSuffix(target, ".") {
			target = target + "."
		}
		err = rc.SetTarget(target)

	case "TXT":
		err = rc.SetTargetTXT(r.Value)

	case "SRV":
		rc.SrvPriority = uint16(r.Priority)
		rc.SrvWeight = uint16(r.Weight)
		rc.SrvPort = uint16(r.Port)
		target := r.Value
		if !strings.HasSuffix(target, ".") {
			target = target + "."
		}
		err = rc.SetTarget(target)

	default:
		err = fmt.Errorf("unsupported record type: %s", r.RecordType)
	}

	return rc, err
}

// recordToLegacy converts a dnscontrol RecordConfig to a UniFi legacy API record.
// Note: This returns a legacyDNSRecord for reference, but actual API calls use recordToLegacyMap.
func recordToLegacy(domain string, rc *models.RecordConfig) (*legacyDNSRecord, error) {
	r := &legacyDNSRecord{
		Enabled:    true,
		Key:        rc.NameFQDN,
		RecordType: rc.Type,
		TTL:        int(rc.TTL),
	}

	switch rc.Type {
	case "A", "AAAA":
		r.Value = rc.GetTargetField()

	case "CNAME", "NS":
		r.Value = strings.TrimSuffix(rc.GetTargetField(), ".")

	case "MX":
		r.Priority = int(rc.MxPreference)
		r.Value = strings.TrimSuffix(rc.GetTargetField(), ".")

	case "TXT":
		r.Value = rc.GetTargetTXTJoined()

	case "SRV":
		r.Priority = int(rc.SrvPriority)
		r.Weight = int(rc.SrvWeight)
		r.Port = int(rc.SrvPort)
		r.Value = strings.TrimSuffix(rc.GetTargetField(), ".")

	default:
		return nil, fmt.Errorf("unsupported record type: %s", rc.Type)
	}

	return r, nil
}

// recordToLegacyMap converts a dnscontrol RecordConfig to a map for API requests.
// UniFi is strict about which fields can be set for each record type.
func recordToLegacyMap(domain string, rc *models.RecordConfig) (map[string]any, error) {
	m := map[string]any{
		"enabled":     true,
		"key":         rc.NameFQDN,
		"record_type": rc.Type,
		"value":       "",
	}

	switch rc.Type {
	case "A":
		m["value"] = rc.GetTargetField()
		// A records can have TTL
		if rc.TTL > 0 {
			m["ttl"] = int(rc.TTL)
		}

	case "AAAA":
		m["value"] = rc.GetTargetField()
		// AAAA records can have TTL
		if rc.TTL > 0 {
			m["ttl"] = int(rc.TTL)
		}

	case "CNAME":
		m["value"] = strings.TrimSuffix(rc.GetTargetField(), ".")
		// CNAME records can have TTL
		if rc.TTL > 0 {
			m["ttl"] = int(rc.TTL)
		}

	case "NS":
		m["value"] = strings.TrimSuffix(rc.GetTargetField(), ".")
		// NS records can have TTL
		if rc.TTL > 0 {
			m["ttl"] = int(rc.TTL)
		}

	case "MX":
		// MX records: only enabled, key, record_type, value, priority allowed
		m["value"] = strings.TrimSuffix(rc.GetTargetField(), ".")
		m["priority"] = int(rc.MxPreference)

	case "TXT":
		// TXT records: only enabled, key, record_type, value allowed
		m["value"] = rc.GetTargetTXTJoined()

	case "SRV":
		// SRV records: enabled, key, record_type, value, priority, weight, port allowed
		m["value"] = strings.TrimSuffix(rc.GetTargetField(), ".")
		m["priority"] = int(rc.SrvPriority)
		m["weight"] = int(rc.SrvWeight)
		m["port"] = int(rc.SrvPort)

	default:
		return nil, fmt.Errorf("unsupported record type: %s", rc.Type)
	}

	return m, nil
}

// getRecordID extracts the UniFi record ID from the Original field.
func getRecordID(rc *models.RecordConfig) string {
	if rc.Original == nil {
		return ""
	}
	if r, ok := rc.Original.(*legacyDNSRecord); ok {
		return r.ID
	}
	return ""
}

// recordKey generates a unique key for a record to help with matching.
func recordKey(rc *models.RecordConfig) string {
	switch rc.Type {
	case "SRV":
		return fmt.Sprintf("%s|%s|%d|%d|%d|%s",
			rc.NameFQDN, rc.Type, rc.SrvPriority, rc.SrvWeight, rc.SrvPort, rc.GetTargetField())
	case "MX":
		return fmt.Sprintf("%s|%s|%d|%s",
			rc.NameFQDN, rc.Type, rc.MxPreference, rc.GetTargetField())
	case "TXT":
		return fmt.Sprintf("%s|%s|%s",
			rc.NameFQDN, rc.Type, rc.GetTargetTXTJoined())
	default:
		return fmt.Sprintf("%s|%s|%s",
			rc.NameFQDN, rc.Type, rc.GetTargetField())
	}
}

// parseSRVLabel parses an SRV label like "_sip._tcp" into service and protocol.
func parseSRVLabel(label string) (service, protocol string) {
	parts := strings.SplitN(label, ".", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return label, ""
}

// formatSRVContent formats SRV record content for display.
func formatSRVContent(priority, weight, port int, target string) string {
	return strconv.Itoa(priority) + " " + strconv.Itoa(weight) + " " + strconv.Itoa(port) + " " + target
}
