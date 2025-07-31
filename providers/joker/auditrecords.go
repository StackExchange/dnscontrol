package joker

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider. If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	var errs []error

	// Supported record types
	supported := map[string]bool{
		"A":     true,
		"AAAA":  true,
		"CAA":   true,
		"CNAME": true,
		"MX":    true,
		"NAPTR": true,
		"NS":    true,
		"SRV":   true,
		"TXT":   true,
	}

	for _, rc := range records {
		// Check if record type is supported
		if !supported[rc.Type] {
			errs = append(errs, fmt.Errorf("joker does not support %s records", rc.Type))
			continue
		}


		// Validate SRV records
		if rc.Type == "SRV" {
			if rc.SrvPort == 0 {
				errs = append(errs, fmt.Errorf("SRV records must have a non-zero port"))
			}
			if rc.GetTargetField() == "" {
				errs = append(errs, fmt.Errorf("SRV records must have a target"))
			}
		}

		// Validate CAA records
		if rc.Type == "CAA" {
			if rc.CaaTag == "" {
				errs = append(errs, fmt.Errorf("CAA records must have a tag"))
			}
			if rc.GetTargetField() == "" {
				errs = append(errs, fmt.Errorf("CAA records must have a value"))
			}
		}

		// Validate NAPTR records
		if rc.Type == "NAPTR" {
			if rc.GetTargetField() == "" {
				errs = append(errs, fmt.Errorf("NAPTR records must have a replacement"))
			}
		}

		// Validate NS records - Joker does not allow custom NS records at apex
		if rc.Type == "NS" && rc.Name == "" {
			// This is an NS record at the apex domain
			// Joker automatically manages apex NS records and does not allow custom ones
			errs = append(errs, fmt.Errorf("joker does not support custom NS records at apex (domain root)"))
		}
	}

	return errs
}
