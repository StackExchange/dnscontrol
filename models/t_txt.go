package models

import (
	"strings"

	"github.com/StackExchange/dnscontrol/v3/pkg/decode"
)

/*

HOW DO TXT RECORDS WORK IN GENERAL:

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
		parsing.  However, usually the format is "quoted like in RFC 1035".
		Package `pkg/decode` provides RFC1035Fields(), QuoteEscapedFields(),
		and QuotedFields().  One of those should work.

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

/*

HOW TO PARSE/SEND TXT STRINGS:

- Step 1: Does the API take/send 1 string or a list of strings?
	- List of strings: Your nativeToRc() should call  SetTargetTXTs(many
		[]string); your Corrections code should copy the individual strings
		from rc.TxtStrings (which is []string).  Once this is implemented
		you are done.
	- 1 string: Go to Step 2.

- Step 2: Go to the provider's website and manually create a TXT
	record.  Does it permit you to enter 1 string or a list of strings?
	- 1 string: Your nativeToRc() should call  SetTargetTXT(s string);
		your Corrections code should read the field using
		rc.GetTargetTXTJoined().  Once this is implemented you are done.
  - List of strings: Go to Step 3.

- Step 3: At this point, we can conclude that the string the API gives
	you needs to be parsed into many separate strings.
	- Your nativeToRc() should call SetTargetTXTs(decode.PARSER(s)); your
		Corrections code should read the field using rc.GetTargetRFC1035Quoted().
	- PARSER should be one of:
		- decode.QuoteEscapedFields() -- Handles strings like: "one" "two" "in\"side"
		- decode.QuotedFields() -- Handles quotes, doesn't allow escaped chars.
		- decode.MiekgDNSFields() -- Uses `miekg/dns`'s parser, which leaves backslashes intact.
		- decode.RFC1035Fields() -- Similar to decode.MiekgDNSFields but properly de-escapes quotes.
		- Otherwise... write your own decoder. Please add it to
			`pkg/decode/decoders.go` if you feel others will find it useful.

*/

// HasFormatIdenticalToTXT returns if a RecordConfig has a format which is
// identical to TXT, such as SPF. For more details, read
// https://tools.ietf.org/html/rfc4408#section-3.1.1
func (rc *RecordConfig) HasFormatIdenticalToTXT() bool {
	return rc.Type == "TXT" || rc.Type == "SPF"
}

// SetTargetTXT sets the TXT fields when there is 1 string.
// The string is stored in the first element of .TxtStrings.
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

// GetTargetTXTJoined returns the TXT target as one string. If it was stored as multiple strings, concatenate them.
func (rc *RecordConfig) GetTargetTXTJoined() string {
	return strings.Join(rc.TxtStrings, "")
}

// GetTargetTXTFlattened255 returns the TXT target as a list of
// strings, each 255-octets or shorter.
func (rc *RecordConfig) GetTargetTXTFlattened255() []string {
	return decode.Flatten255(rc.TxtStrings)
}

// There is no GetTargetTXTfromRFC1025Quoted(). Use GetTargetRFC1035Quoted()

// SetTargetTXTString sets the TXT strings after calling decode.QuotedFields().
//
// Deprecated: This function has a confusing name. Use
// SetTargetTXTs(decode.PARSER(s)) where PARSER is one of the provided
// parsers in pkg/decode or write your own.
//
func (rc *RecordConfig) SetTargetTXTString(s string) error {
	ts, err := decode.QuotedFields(s)
	if err != nil {
		return err
	}
	return rc.SetTargetTXTs(ts)
}

// SetTargetTXTQuotedFields sets a TXT target after decoding s.
func (rc *RecordConfig) SetTargetTXTQuotedFields(s string) error {
	ts, err := decode.QuotedFields(s)
	if err != nil {
		return err
	}
	return rc.SetTargetTXTs(ts)
}

// SetTargetTXTQuoteEscapedFields sets a TXT target after decoding s.
func (rc *RecordConfig) SetTargetTXTQuoteEscapedFields(s string) error {
	ts, err := decode.QuoteEscapedFields(s)
	if err != nil {
		return err
	}
	return rc.SetTargetTXTs(ts)
}

// SetTargetTXTMiekgDNSFields sets a TXT target after decoding s.
func (rc *RecordConfig) SetTargetTXTMiekgDNSFields(s string) error {
	ts, err := decode.MiekgDNSFields(s)
	if err != nil {
		return err
	}
	return rc.SetTargetTXTs(ts)
}

// SetTargetTXTRFC1035Fields sets a TXT target after decoding s.
func (rc *RecordConfig) SetTargetTXTRFC1035Fields(s string) error {
	ts, err := decode.RFC1035Fields(s)
	if err != nil {
		return err
	}
	return rc.SetTargetTXTs(ts)
}
