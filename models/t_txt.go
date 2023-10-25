package models

import (
	"fmt"
	"strings"
)

/*
Sadly many providers handle TXT records in strange and non-compliant
ways.  DNSControl has to handle all of them.  Over the years we've
tried many things.  This explain the current state of the code.

DNSControl stores the TXT record target as a single string of any length.
Providers take care of any splitting, excaping, or quoting.

NOTE: Older versions of DNSControl stored the TXT record as
represented by the provider, which could be a single string, a series
of smaller strings, or a single string that is quoted/escaped.  This
created tons of edge-cases and other distractions.

If a provider doesn't support certain charactors in a TXT record, use
the providers/$PROVIDER/auditrecords.go file to indicate this.
DNSControl uses this information to warn users of unsupporrted input,
and to skip related integration tests.

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

	return rc.SetTarget(s)
}

// SetTargetTXTs sets the TXT fields when there are many strings.
// The individual strings are stored in .TxtStrings, and joined to make .Target.
func (rc *RecordConfig) SetTargetTXTs(s []string) error {
	return rc.SetTargetTXT(strings.Join(s, ""))
}

// GetTargetTXTJoined returns the TXT target as one string. If it was stored as multiple strings, concatenate them.
// Deprecated: GetTargetTXTJoined is deprecated. Use GetTargetField()
func (rc *RecordConfig) GetTargetTXTJoined() string {
	return rc.GetTargetField()
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
//
// Deprecated: GetTargetTXTJoined is deprecated. ...or should be.
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
