package models

import (
	"strings"

	"github.com/StackExchange/dnscontrol/v4/pkg/txtutil"
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
func (rc *RecordConfig) SetTargetTXT(s string) error {
	if rc.Type == "" {
		rc.Type = "TXT"
	} else if !rc.HasFormatIdenticalToTXT() {
		panic("assertion failed: SetTargetTXT called when .Type is not TXT or compatible type")
	}

	return rc.SetTarget(s)
}

// SetTargetTXTs sets the TXT fields when there are many strings. They are stored concatenated.
func (rc *RecordConfig) SetTargetTXTs(s []string) error {
	return rc.SetTargetTXT(strings.Join(s, ""))
}

// GetTargetTXTJoined returns the TXT target as one string.
func (rc *RecordConfig) GetTargetTXTJoined() string {
	return rc.target
}

// GetTargetTXTChunked255 returns the TXT target as a list of strings, 255 octets each with the remainder on the last string.
func (rc *RecordConfig) GetTargetTXTChunked255() []string {
	return txtutil.ToChunks(rc.target)
}
