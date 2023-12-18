package dnsimple

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("MX", rejectif.MxNull) // Last verified 2023-03

	a.Add("TXT", rejectif.TxtLongerThan255) // Last verified 2023-03

	a.Add("TXT", rejectif.TxtHasTrailingSpace) // Last verified 2023-03

	a.Add("TXT", rejectif.TxtHasUnpairedDoubleQuotes) // Last verified 2023-03

	a.Add("TXT", rejectif.TxtIsEmpty) // Last verified 2023-03

	return a.Audit(records)
}
