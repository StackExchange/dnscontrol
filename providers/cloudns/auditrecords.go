package cloudns

import (
	"github.com/StackExchange/dnscontrol/v3/models"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a = rejectif.Auditor{}

	a.Add("TXT", rejectif.TxtHasBackticks) // needed as of 2021-03-01

	a.Add("TXT", rejectif.TxtIsEmpty) // needed as of 2021-03-01

	a.Add("TXT", rejectif.TxtHasTrailingSpace) // needed as of 2021-03-01

	a.Add("TXT", rejectif.TxtHasDoubleQuotes) // needed as of 2021-03-01

	a.Add("TXT", rejectif.TxtHasMultipleStrings) // needed as of 2021-03-01

	return audits.Audit()
}
