package joker

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider. If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	// Joker does not support custom NS records at apex (domain root)
	// Joker automatically manages apex NS records
	a.Add("NS", rejectif.NsAtApex) // Last verified 2025-01-31

	// Joker has round-trip issues with TXT records containing unbalanced quotes
	a.Add("TXT", rejectif.TxtHasUnpairedDoubleQuotes) // Last verified 2025-01-31

	// Joker has round-trip issues with TXT records containing backslashes
	a.Add("TXT", rejectif.TxtHasBackslash) // Last verified 2025-01-31

	// SRV records must have valid port and target
	a.Add("SRV", rejectif.SrvHasZeroPort) // Last verified 2025-01-31
	a.Add("SRV", rejectif.SrvHasEmptyTarget) // Last verified 2025-01-31

	// CAA records must have valid tag and target
	a.Add("CAA", rejectif.CaaHasEmptyTag) // Last verified 2025-01-31
	a.Add("CAA", rejectif.CaaHasEmptyTarget) // Last verified 2025-01-31

	// NAPTR records must have a replacement
	a.Add("NAPTR", rejectif.NaptrHasEmptyTarget) // Last verified 2025-01-31

	return a.Audit(records)
}
