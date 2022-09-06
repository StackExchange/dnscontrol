package digitalocean

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("CAA", rejectif.CaaTargetHasSemicolon) // Last verified 2021-03-01

	a.Add("CAA", rejectif.CaaTargetContainsWhitespace) // Last verified xxxx-xx-xx

	a.Add("TXT", MaxLengthDO) // Last verified 2021-03-01

	a.Add("TXT", rejectif.TxtHasDoubleQuotes) // Last verified 2021-03-01
	// Double-quotes not permitted in TXT strings. I have a hunch that
	// this is due to a broken parser on the DO side.

	return a.Audit(records)
}

// MaxLengthDO returns and error if the string is longer than
// permitted by DigitalOcean. Sadly their length limit is
// undocumented. This is a guess.
func MaxLengthDO(rc *models.RecordConfig) error {
	// The total length of all strings can't be longer than 512; and in
	// reality must be shorter due to sloppy validation checks.
	// https://github.com/StackExchange/dnscontrol/issues/370

	// DigitalOcean's TXT record implementation checks size limits
	// wrong.  RFC 1035 Section 3.3.14 states that each substring can be
	// 255 octets, and there is no limit on the number of such
	// substrings, aside from the usual packet length limits.  DO's
	// implementation restricts the total length to be 512 octets,
	// including the quotes, backlashes used for escapes, spaces between
	// substrings.
	// In other words, they're doing the checking on the API protocol
	// encoded data instead of on on the resulting TXT record.  Sigh.

	if len(rc.GetTargetField()) > 509 {
		return fmt.Errorf("encoded txt too long")
	}
	// FIXME(tlim): Try replacing GetTargetField() with (2 + (3*len(rc.TxtStrings) - 1))

	return nil
}
