package msdns

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("MX", rejectif.MxNull) // Last verified 2023-02-02

	a.Add("SRV", rejectif.SrvHasNullTarget) // Last verified 20-0212-28

	a.Add("TXT", rejectif.TxtHasBackticks) // Last verified 2023-02-02

	a.Add("TXT", rejectif.TxtHasDoubleQuotes) // Last verified 2023-02-02

	a.Add("TXT", rejectif.TxtHasMultipleSegments) // Last verified 2023-02-02

	a.Add("TXT", rejectif.TxtHasSegmentLen256orLonger) // Last verified 2023-02-02

	a.Add("TXT", rejectif.TxtHasSingleQuotes) // Last verified 2023-02-02

	a.Add("TXT", rejectif.TxtIsEmpty) // Last verified 2023-02-02

	a.Add("TXT", rejectif.TxtIsExactlyLen255) // Last verified 2023-02-02

	a.Add("TXT", rejectif.TxtIsExactlyLen255) // Last verified 2023-02-02

	return a.Audit(records)
}
