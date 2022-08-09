package cloudflare

import "github.com/StackExchange/dnscontrol/v3/models"

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a = rejectif.Auditor{}

	a.Add("TXT", rejectif.TxtHasMultipleStrings) // needed as of 2022-06-18

	a.Add("TXT", rejectif.TxtHasTrailingSpace) // needed as of 2022-06-18

	a.Add("TXT", rejectif.TxtIsEmpty) // needed as of 2022-06-18

	return a.Audit()
}
