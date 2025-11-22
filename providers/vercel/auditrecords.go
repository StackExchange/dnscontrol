package vercel

import (
	"errors"

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

	// last verified 2025-11-22
	// bad_request - invalid_value - The specified value is not a fully qualified domain name.
	a.Add("CAA", rejectif.CaaHasEmptyTarget)
	a.Add("CAA", rejectifCaaTargetIsSemicolon)

	// last verified 2025-11-22
	// Vercel misidentified extra fields in CAA record `0 issue letsencrypt.org; validationmethods=dns-01; accounturi=https://acme-v02.api.letsencrypt.org/acme/acct/1234`
	// as "cansignhttpexchanges", and add extra incorrect validation on the value
	// let's ignore all whitespace for now, i should report this to Vercel though, as
	// it uses NS1 as its provder and NS1 definitly allows it.
	//
	// invalid_value - Unexpected "cansignhttpexchanges" value.
	a.Add("CAA", rejectif.CaaTargetContainsWhitespace)

	return a.Audit(records)
}

func rejectifCaaTargetIsSemicolon(rc *models.RecordConfig) error {
	if rc.GetTargetField() == ";" {
		return errors.New("caa target cannot be ';'")
	}
	return nil
}
