package hexonet

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("TXT", rejectif.TxtIsEmpty) // Last verified 2021-10-01

	a.Add("TXT", rejectif.TxtHasDoubleQuotes) // Last verified 2023-11-30

	a.Add("SRV", rejectif.SrvHasNullTarget) // Last verified 2020-12-28

	return a.Audit(records)
}
