package models

import (
	"fmt"
	"strings"
)

// SetTargetMX sets the MX fields.
func (rc *RecordConfig) SetTargetMX(pref uint16, target string) error {
	rc.MxPreference = pref
	if err := rc.SetTarget(target); err != nil {
		return err
	}
	if rc.Type == "" {
		rc.Type = "MX"
	}
	if rc.Type != "MX" {
		panic("assertion failed: SetTargetMX called when .Type is not MX")
	}
	return rc.PopulateMXFields(pref, target, nil, "")
}

// SetTargetMXStrings is like SetTargetMX but accepts strings.
func (rc *RecordConfig) SetTargetMXStrings(pref, target string) error {
	return PopulateMXRaw(rc, []string{pref, target}, nil, "")
}

// SetTargetMXString is like SetTargetMX but accepts one big string.
func (rc *RecordConfig) SetTargetMXString(s string) error {
	part := strings.Fields(s)
	if len(part) != 2 {
		return fmt.Errorf("MX value does not contain 2 fields: (%#v)", s)
	}
	return rc.SetTargetMXStrings(part[0], part[1])
}
