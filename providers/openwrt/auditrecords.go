package openwrt

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rejectif"
)

var supportedRTypes = map[string]struct{}{
	"A":     {},
	"AAAA":  {},
	"CNAME": {},
	"MX":    {},
	"SRV":   {},
}

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	// MX records cannot have null/empty target
	a.Add("MX", rejectif.MxNull)

	// SRV records cannot have null target
	a.Add("SRV", rejectif.SrvHasNullTarget)

	// Start with auditor errors
	var errors []error
	errors = append(errors, a.Audit(records)...)

	// Check for unsupported record types
	for _, r := range records {
		if _, ok := supportedRTypes[r.Type]; !ok {
			errors = append(errors, fmt.Errorf("record type %q is not supported by OpenWrt", r.Type))
		}

		// OpenWrt doesn't support wildcard CNAMEs
		if r.Type == "CNAME" && r.GetLabel() == "*" {
			errors = append(errors, fmt.Errorf("OpenWrt does not support wildcard CNAME records"))
		}
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}
