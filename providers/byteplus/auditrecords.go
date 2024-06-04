package byteplus

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("TXT", rejectif.TxtIsEmpty)
	a.Add("TXT", rejectif.TxtHasDoubleQuotes)
	a.Add("TXT", rejectif.TxtHasSingleQuotes)
	a.Add("TXT", rejectif.TxtHasUnpairedDoubleQuotes)
	a.Add("MX", rejectif.MxNull)
	a.Add("MX", rejectif.MxPriorityMoreThan100)
	a.Add("SRV", rejectif.SrvHasNullTarget)
	a.Add("CAA", rejectif.CaaTargetContainsWhitespace)

	return a.Audit(records)
}
