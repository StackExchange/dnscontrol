package cscglobal

import (
	"github.com/StackExchange/dnscontrol/v3/models"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a = rejectif.Auditor{}

	a.Add("TXT", rejectif.TxtNoMultipleStrings) // Still needed as of 2022-06-10
	a.Add("TXT", rejectif.TxtNoTrailingSpace)   // Still needed as of 2022-06-10
	a.Add("TXT", rejectif.TxtNotEmpty)          // Still needed as of 2022-06-10

	return a.Audit()
}

/* How To Write Providers:

Each test should be encapsulated in a function that can be tested
individually.  If the test is of general use, add it to the
rejectif module.

The "Still needed as of" comment logs the last time we verified this
test was needed.  Sometimes companies change their API.  Once a year,
try removing tests one at a time to verify they are still needed.

*/
