package vercel

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	// last verified 2025-11-22
	// vercel does not support custom NS records at apex (domain root)
	// vercel automatically manages apex NS records
	// attempted to set one will result in "invalid_name - Cannot set NS records at the root level. Only subdomain NS records are supported"
	a.Add("NS", rejectif.NsAtApex)

	// last verified 2025-11-22
	// bad_request - Invalid request: The specified value is not a fully qualified domain name.
	a.Add("MX", rejectif.MxNull)

	// last verified 2025-11-22
	// bad_request - Invalid request: missing required property `value`.
	a.Add("TXT", rejectif.TxtIsEmpty)

	return a.Audit(records)
}
