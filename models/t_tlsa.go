package models

import (
	"fmt"
	"strings"
)

// SetTargetTLSA sets the TLSA fields.
func (rc *RecordConfig) SetTargetTLSA(usage, selector, matchingtype uint8, target string) {
	rc.TlsaUsage = usage
	rc.TlsaSelector = selector
	rc.TlsaMatchingType = matchingtype
	rc.Target = target
	if rc.Type == "" {
		rc.Type = "TLSA"
	}
	if rc.Type != "TLSA" {
		panic("SetTargetTLSA called when .Type is not TLSA")
	}
}

// SetTargetTLSAStrings is like SetTargetTLSA but accepts strings.
func (rc *RecordConfig) SetTargetTLSAStrings(usage, selector, matchingtype, target string) {
	rc.SetTargetTLSA(atou8(usage), atou8(selector), atou8(matchingtype), target)
}

// SetTargetTLSAString is like SetTargetTLSA but accepts one big string.
func (rc *RecordConfig) SetTargetTLSAString(s string) {
	part := strings.Fields(s)
	if len(part) != 4 {
		panic(fmt.Errorf("TLSA value %#v contains too many fields", s))
	}
	rc.SetTargetTLSAStrings(part[0], part[1], part[2], part[3])
}
