package digitalocean

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/recordaudit"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a = rejectif.Auditor{}

	//a.Add("CAA", recordaudit.CaaNoSemicolIssue) // Still needed as of 2021-03-01

	a.Add("TXT", MaxLengthDO) // Still needed as of 2021-03-01

	a.Add("TXT", recordaudit.TxtNoDoubleQuotes) // Still needed as of 2021-03-01
	// Double-quotes not permitted in TXT strings. I have a hunch that
	// this is due to a broken parser on the DO side.

	return a.Audit()
}
