package msdns

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("TXT", rejectif.TxtHasBackticks) // Last verified 2021-03-01

	a.Add("TXT", rejectif.TxtHasDoubleQuotes) // Last verified 2021-03-01

	a.Add("TXT", rejectif.TxtHasMultipleSegments) // Last verified 2021-03-01

	a.Add("TXT", rejectif.TxtHasSingleQuotes) // Last verified 2021-03-01

	a.Add("TXT", rejectif.TxtHasSegmentLen256orLonger)

	a.Add("TXT", rejectif.TxtIsEmpty) // Last verified 2021-03-01

	a.Add("SRV", rejectif.SrvHasNullTarget) // Last verified 2020-12-28

	return a.Audit(records)
}
