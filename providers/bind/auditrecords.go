package bind

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("TXT", rejectif.TxtHasDoubleQuotes) // Last verified 2023-06-12
	// This is supported by the provider but I'm too lazy to get the
	// quoting right at this time.  When we overhaul how TXT records are
	// done, this will be easy to fix. --tlim 2023-06-15

	return a.Audit(records)
}
