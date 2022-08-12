package vultr

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("TXT", rejectif.TxtHasDoubleQuotes) // Last verified 2021-03-02
	// Needs investigation. Could be a dnscontrol issue or
	// the provider doesn't support double quotes.

	a.Add("TXT", rejectif.TxtHasMultipleSegments)

	return a.Audit(records)
}
