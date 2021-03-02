package digitalocean

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/recordaudit"
)

// AuditRecords returns an error if any records are not
// supportable by this provider.
func AuditRecords(records []*models.RecordConfig) error {

	// TODO(tlim): Audit CAA records.
	// "Semicolons not supported in issue/issuewild fields.", "https://www.digitalocean.com/docs/networking/dns/how-to/create-caa-records"),
	//	Users are warned about these limits in docs/_providers/digitalocean.md

	if err := MaxLengthDO(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-01

	// Double-quotes not permitted in TXT strings. I have a hunch that
	// this is due to a broken parser on the DO side.
	if err := recordaudit.TxtNoDoubleQuotes(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-01

	return nil
}

// MaxLengthDO returns and error if the strings are longer than
// permitted by DigitalOcean. Sadly their length limit is
// undocumented. This is a guess.
func MaxLengthDO(records []*models.RecordConfig) error {
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

	for _, rc := range records {

		if rc.HasFormatIdenticalToTXT() { // TXT and similar:
			if len(rc.GetTargetField()) > 509 {
				return fmt.Errorf("encoded txt too long")
			}
		}

	}
	return nil
}
