package gidinet

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider. If all records are supported,
// an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("MX", rejectif.MxNull) // MX priority 0 is allowed (means highest priority)

	a.Add("TXT", rejectif.TxtHasDoubleQuotes) // Gidinet doesn't support quotes in TXT
	a.Add("TXT", rejectif.TxtIsEmpty)         // Empty TXT records not allowed
	a.Add("TXT", rejectif.TxtHasBackticks)    // Backticks not supported

	a.Add("SRV", rejectif.SrvHasNullTarget) // SRV must have a target

	return a.Audit(records)
}
