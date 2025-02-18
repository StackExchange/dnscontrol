package models

import (
	"strings"
)

// SetTargetMX sets the MX fields.
func (rc *RecordConfig) SetTargetMX(preference uint16, mx string) error {
	rc.Type = "MX"
	return RecordUpdateFields(rc, MX{Preference: preference, Mx: mx}, nil)
}

// SetTargetMXStrings is like SetTargetMX but accepts strings.
func (rc *RecordConfig) SetTargetMXStrings(pref, target string) error {
	rdata, err := ParseMX([]string{rc.Name, pref, target}, "", "")
	if err != nil {
		return err
	}
	return RecordUpdateFields(rc, rdata, nil)
}

// SetTargetMXString is like SetTargetMX but accepts one big string.
func (rc *RecordConfig) SetTargetMXString(s string) error {
	part := strings.Fields(s)
	return rc.SetTargetMXStrings(part[0], part[1])
}
