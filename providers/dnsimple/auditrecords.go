package dnsimple

import (
	"github.com/StackExchange/dnscontrol/v3/models"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a = rejectif.Auditor{}

	a.Add("TXT", rejectif.TxtHasMultipleStrings) // Still needed as of 2022-07
	//TODO(onlyhavecans) I think we can support multiple strings.

	a.Add("TXT", rejectif.TxtHasTrailingSpace) // Still needed as of 2022-07

	a.Add("TXT", rejectif.TxtIsEmpty) // Still needed as of 2022-07

	a.Add("TXT", rejectif.TxtHasUnpairedDoubleQuotes) // Still needed as of 2022-07

	return a.Audit()
}
