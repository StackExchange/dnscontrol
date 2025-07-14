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

	for _, rc := range records {
		if rc.Type == "ALIAS" {
			errs = append(errs, fmt.Errorf("joker does not support ALIAS records"))
		}
		if rc.Type == "DS" {
			errs = append(errs, fmt.Errorf("joker does not support DS records"))
		}
		if rc.Type == "DNSKEY" {
			errs = append(errs, fmt.Errorf("joker does not support DNSKEY records"))
		}
		if rc.Type == "HTTPS" {
			errs = append(errs, fmt.Errorf("joker does not support HTTPS records"))
		}
		if rc.Type == "LOC" {
			errs = append(errs, fmt.Errorf("joker does not support LOC records"))
		}
		if rc.Type == "PTR" {
			errs = append(errs, fmt.Errorf("joker does not support PTR records"))
		}
		if rc.Type == "SOA" {
			errs = append(errs, fmt.Errorf("joker does not support SOA records"))
		}
		if rc.Type == "SSHFP" {
			errs = append(errs, fmt.Errorf("joker does not support SSHFP records"))
		}
		if rc.Type == "SVCB" {
			errs = append(errs, fmt.Errorf("joker does not support SVCB records"))
		}
		if rc.Type == "TLSA" {
			errs = append(errs, fmt.Errorf("joker does not support TLSA records"))
		}

		// Check TTL minimum - NAPTR and SVC records can have TTL=0, others need >= 300
		if rc.TTL != 0 && rc.TTL < 300 {
			if rc.Type != "NAPTR" && rc.Type != "SVC" {
				errs = append(errs, fmt.Errorf("joker requires TTL to be 300 or higher (except NAPTR and SVC which can be 0)"))
			}
		}

		// Validate MX records
		if rc.Type == "MX" {
			if rc.MxPreference == 0 {
				errs = append(errs, fmt.Errorf("MX records must have a preference value"))
			}
		}

		// Validate SRV records
		if rc.Type == "SRV" {
			if rc.SrvPort == 0 {
				errs = append(errs, fmt.Errorf("SRV records must have a port"))
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
	}

	return errs
}
