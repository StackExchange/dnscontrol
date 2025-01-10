package models

import (
	"fmt"
	"strings"
)

// SetTargetMX sets the MX fields.
func (rc *RecordConfig) SetTargetMX(pref uint16, target string) error {
	return rc.PopulateFieldsMX(pref, target, nil, "")
}

// SetTargetMXStrings is like SetTargetMX but accepts strings.
func (rc *RecordConfig) SetTargetMXStrings(pref, target string) error {
	return PopulateFromRawMX(rc, []string{rc.Name, pref, target}, nil, "")
}

// SetTargetMXString is like SetTargetMX but accepts one big string.
func (rc *RecordConfig) SetTargetMXString(s string) error {
	part := strings.Fields(s)
	if len(part) != 2 {
		return fmt.Errorf("MX value does not contain 2 fields: (%#v)", s)
	}
	return rc.SetTargetMXStrings(part[0], part[1])
}
