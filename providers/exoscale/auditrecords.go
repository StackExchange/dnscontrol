package exoscale

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("CAA", rejectif.CaaTargetContainsWhitespace) // Last verified 2022-07-11

	a.Add("MX", rejectif.MxNull) // Last verified 2022-07-11

	a.Add("SRV", rejectif.SrvHasNullTarget) // Last verified 2020-12-28

	a.Add("TXT", rejectif.TxtHasUnpairedDoubleQuotes) // Last verified 2022-09-14

	return a.Audit(records)
}
