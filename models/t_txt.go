package models

import (
	"fmt"
	"strings"
)

/*
Sadly many providers handle TXT records in strange and non-compliant
ways.  DNSControl has to handle all of them.  Over the years we've
tried many things.  This explain the current state of the code.

What are some of these variations?

* The RFCs say that a TXT record is a series of strings, each 255-octets
  or fewer.  Yet, most provider APIs only support a single string which
  is split into 255-octetl chunks behind the scenes.  Some only support
  a single string that is 255-octets or less.

* The RFCs don't say much about the content of the strings.  Some
  providers accept any octet, some only accept ASCII-printable chars,
  some get confused by TXT records that include backticks, quotes, or
  whitespace at the end of the string.

DNSControl has tried many different ways to handle all these
variations over the years. This is what we found works best:

Principle 1. Store the string as the user input it.

DNSControl stores the string as the user specified in dnsconfig.js.
The user can specify a string of any length, or many individual
strings of any length.

No matter how the user presented the data in dnsconfig.js, the data is
stored as a list of strings (RecordConfig.TxtStrings []string).  If
they input 1 string, the list has one element. If the user input many
individual strings, the list is copied into .TxtStrings.

When we store the data in .TxtStrings there is no length checking. The data is not manipulated.

Principle 2. When downloading zone records, receive the data as appropriate.

When the API returns a TXT record, the provider's code must properly
store it in the .TxtStrings field of RecordConfig.

We've found most APIs return TXT strings in one of three ways:

  * The API returns a single string: use RecordConfig.SetTargetTXT().
  * The API returns multiple strings: use RecordConfig.SetTargetTXTs().
	* (THIS IS RARE) The API returns a single string that must be parsed
		into multiple strings: The provider is responsible for the
		parsing.  However, usually the format is "quoted like in RFC 1035"
		which is vague, but we've implemented it as
		RecordConfig.SetTargetTXTfromRFC1035Quoted().

If the format is something else, please write the parser as a separate
function and write unit tests based on actual data received from the
API.

Principle 3. When sending TXT records to the API, send what the API expects.

The provider's code must decide how to take the list of strings in
.TxtStrings and present them to the API.

Most providers fall into one of these categories:

	* If the API expects one long string, the provider code joins all
	  the smaller strings and sends one big string.  Use the helper
	  function RecordConfig.GetTargetTXTJoined()
  * If the API expects many strings of any size, the provider code
	  sends the individual strings. Those strings are accessed as
	  the array RecordConfig.TxtStrings
	* (THIS IS RARE) If the API expects multiple strings to be sent as
	  one long string, quoted RFC 1025-style, call
	  RecordConfig.GetTargetRFC1035Quoted() and send that string.

Note: If the API expects many strings, each 255-octets or smaller, the
provider code must split the longer strings into smaller strings.  The
helper function txtutil.SplitSingleLongTxt(dc.Records) will iterate
over all TXT records and split out any strings longer than 255 octets.
Call this once in GetDomainCorrections().  (Yes, this violates
Principle 1, but we decided it is best to do it once, than provide a
getter that would re-split the strings on every call.)

Principle 4. Providers can communicate back to DNSControl strings they can't handle.

As mentioned before, some APIs reject TXT records for various reasons:
Illegal chars, whitespace at the end, etc.  We can't make a flag for
every variation.  Instead we call the provider's AuditRecords()
function and it reports if there are any records that it can't
process.

We've provided many helper functions to make this easier.  Look at any
of the providers/.../auditrecord.go` files for examples.

The integration tests call AuditRecords() to skip any tests that we
know will fail.  If one of the integration tests is failing, it is
often better to update AuditRecords() than to try to figure out why,
for example, the provider doesn't support backticks in strings.  Don't
spend a lot of effort trying to fix situations that are rare or will
not appear in real-world situations.

Companies do update their APIs occasionally. You might want to try
eliminating the checks one at a time to see if the API has improved.
Don't feel obligated to do this more than once a year.

Conclusion:

When we follow these 4 principles, and stick with the helper functions
provided, we're able to handle all the variations.

*/

// HasFormatIdenticalToTXT returns if a RecordConfig has a format which is
// identical to TXT, such as SPF. For more details, read
// https://tools.ietf.org/html/rfc4408#section-3.1.1
func (rc *RecordConfig) HasFormatIdenticalToTXT() bool {
	return rc.Type == "TXT" || rc.Type == "SPF"
}

// SetTargetTXT sets the TXT fields when there is 1 string.
// The string is stored in .Target, and split into 255-octet chunks
// for .TxtStrings.
func (rc *RecordConfig) SetTargetTXT(s string) error {
	if rc.Type == "" {
		rc.Type = "TXT"
	} else if !rc.HasFormatIdenticalToTXT() {
		panic("assertion failed: SetTargetTXT called when .Type is not TXT or compatible type")
	}

	rc.TxtStrings = []string{s}
	rc.SetTarget(rc.zoneFileQuoted())
	return nil
}

// SetTargetTXTs sets the TXT fields when there are many strings.
// The individual strings are stored in .TxtStrings, and joined to make .Target.
func (rc *RecordConfig) SetTargetTXTs(s []string) error {
	if rc.Type == "" {
		rc.Type = "TXT"
	} else if !rc.HasFormatIdenticalToTXT() {
		panic("assertion failed: SetTargetTXTs called when .Type is not TXT or compatible type")
	}

	rc.TxtStrings = s
	rc.SetTarget(rc.zoneFileQuoted())
	return nil
}

// GetTargetTXTJoined returns the TXT target as one string.
func (rc *RecordConfig) GetTargetTXTJoined() string {
	return strings.Join(rc.TxtStrings, "")
}

// GetTargetTXTSegmented returns the TXT target as 255-octet segments, with the remainder in the last segment.
func (rc *RecordConfig) GetTargetTXTSegmented() []string {
	return splitChunks(strings.Join(rc.TxtStrings, ""), 255)
}

// GetTargetTXTSegmentCount returns the number of 255-octet segments required to store TXT target.
func (rc *RecordConfig) GetTargetTXTSegmentCount() int {
	var total int
	for i := range rc.TxtStrings {
		total = len(rc.TxtStrings[i])
	}
	segs := total / 255 // integer division, decimals are truncated
	if (total % 255) > 0 {
		return segs + 1
	}
	return segs
}

func splitChunks(buf string, lim int) []string {
	var chunk string
	chunks := make([]string, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:])
	}
	return chunks
}

// SetTargetTXTfromRFC1035Quoted parses a series of quoted strings
// and sets .TxtStrings based on the result.
// Note: Most APIs do notThis is rarely used. Try using SetTargetTXT() first.
// Ex:
//
//	"foo"        << 1 string
//	"foo bar"    << 1 string
//	"foo" "bar"  << 2 strings
//	foo          << error. No quotes! Did you intend to use SetTargetTXT?
func (rc *RecordConfig) SetTargetTXTfromRFC1035Quoted(s string) error {
	if s != "" && s[0] != '"' {
		// If you get this error, it is likely that you should use
		// SetTargetTXT() instead of SetTargetTXTfromRFC1035Quoted().
		return fmt.Errorf("non-quoted string used with SetTargetTXTfromRFC1035Quoted: (%s)", s)
	}
	many, err := ParseQuotedFields(s)
	if err != nil {
		return err
	}
	return rc.SetTargetTXTs(many)
}

// There is no GetTargetTXTfromRFC1025Quoted(). Use GetTargetRFC1035Quoted()
