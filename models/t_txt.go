package models

/*
Sadly many providers handle TXT records in strange and non-compliant ways.

The variations we've seen:

* a TXT record is a list of strings, each 255-octets or fewer.
* a TXT record is a list of strings, all but the last will be 255 octets in length.
* a TXT record is a string less than 256 octets.
* a TXT record is any length, but we'll split it into 255-octet chunks behind the scenes.

DNSControl stores the string as the user specified, then lets the
provider work out how to handle the given input. There are two
opportunties to work with the data:

1. The provider's AuditRecords() function is called to permit
the provider to return an error if it won't be able to handle the
contents. For example, it might detect that the string contains a char
the provider doesn't support (for example, a backtick). This auditing
is done without any communication to the provider's API. This allows
such errors to be detected at the "dnscontrol check" stage.  2. When
performing corrections (GetDomainCorrections()), the provider can slice
and dice the user's input however they want.

* If the user input is a list of strings:
  * The strings are stored in RecordConfig.TxtStrings
* If the user input is a single string:
  * The strings are stored in RecordConfig.TxtStrings[0]

In both cases, the .Target stores a string that can be used in error
messages and other UI messages. This string should not be used by the
provider in any other way because it has been modified from the user's
original input.

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

// SetTargetTXTString is like SetTargetTXT but accepts one big string,
// which must be parsed into one or more strings based on how it is quoted.
// Ex: foo             << 1 string
//     foo bar         << 1 string
//     "foo" "bar"     << 2 strings
func (rc *RecordConfig) SetTargetTXTString(s string) error {
	return rc.SetTargetTXTs(ParseQuotedTxt(s))
}
