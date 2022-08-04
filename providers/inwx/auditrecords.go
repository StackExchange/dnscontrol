package inwx

import "github.com/StackExchange/dnscontrol/v3/models"

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a = rejectif.Auditor{}

	a.Add("TXT", rejectif.TxtNoBackticks) // Still needed as of 2021-03-01

	a.Add("TXT", rejectif.TxtNoStringsExactlyLen255) // Still needed as of 2021-03-01

	a.Add("TXT", rejectif.TxtNoTrailingSpace) // Still needed as of 2021-03-01

	a.Add("TXT", rejectif.TxtNotEmpty) // Still needed as of 2021-03-01

	return a.Audit()
}
