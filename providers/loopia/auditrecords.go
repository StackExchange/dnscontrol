package loopia

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("TXT", rejectif.TxtIsEmpty) // Last verified 2023-03-10: Loopia returns 404

	// Loopias TXT length limit appears to be 450 octets
	a.Add("TXT", rejectif.TxtLongerThan(450)) // Last verified 2023-03-10

	a.Add("MX", rejectif.MxNull) // Last verified 2023-03-23

	return a.Audit(records)
}
