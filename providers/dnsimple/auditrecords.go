package dnsimple

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("MX", rejectif.MxNull) // Last verified 2023-03

	a.Add("TXT", rejectif.TxtHasMultipleSegments) // Last verified 2023-03
	// TODO(onlyhavecans) we can support this, but it needs more work

	a.Add("TXT", rejectif.TxtHasTrailingSpace) // Last verified 2023-03

	a.Add("TXT", rejectif.TxtHasUnpairedDoubleQuotes) // Last verified 2023-03

	a.Add("TXT", rejectif.TxtIsEmpty) // Last verified 2023-03

	return a.Audit(records)
}
