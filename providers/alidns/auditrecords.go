package alidns

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("MX", rejectif.MxNull)                   // Last verified at 2025-12-03
	a.Add("TXT", rejectif.TxtIsEmpty)              // Last verified at 2025-12-03
	a.Add("TXT", rejectif.TxtLongerThan(512))      // Last verified at 2025-12-03: 511 bytes OK, 764 bytes failed
	a.Add("TXT", rejectif.TxtHasDoubleQuotes)      // Last verified at 2025-12-03: Alibaba strips quotes
	a.Add("TXT", rejectif.TxtHasTrailingSpace)     // Last verified at 2025-12-03: Alibaba strips trailing spaces
	a.Add("TXT", rejectif.TxtHasUnpairedBackslash) // Last verified at 2025-12-03: Alibaba mishandles odd backslashes
	return a.Audit(records)
}
