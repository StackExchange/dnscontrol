package models

import (
	"fmt"
	"strings"
)

// SetTargetMX sets the MX fields.
func (rc *RecordConfig) SetTargetMX(pref uint16, target string) {
	rc.MxPreference = pref
	rc.Target = target
	if rc.Type == "" {
		rc.Type = "MX"
	}
	if rc.Type != "MX" {
		panic("SetTargetMX called when .Type is not MX")
	}
}

// SetTargetMXStrings is like SetTargetMX but accepts strings.
func (rc *RecordConfig) SetTargetMXStrings(pref, target string) {
	rc.SetTargetMX(atou16(pref), target)
}

// SetTargetMXString is like SetTargetMX but accepts one big string.
func (rc *RecordConfig) SetTargetMXString(s string) {
	part := strings.Fields(s)
	if len(part) != 2 {
		panic(fmt.Errorf("MX value %#v contains too many fields", s))
	}
	rc.SetTargetMXStrings(part[0], part[1])
}
