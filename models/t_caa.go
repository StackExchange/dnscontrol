package models

import (
	"fmt"
	"strconv"
	"strings"
)

// SetTargetCAA sets the CAA fields.
func (rc *RecordConfig) SetTargetCAA(flag uint8, tag string, target string) error {
	rc.CaaTag = tag
	rc.CaaFlag = flag
	rc.SetTarget(target)
	if rc.Type == "" {
		rc.Type = "CAA"
	}
	if rc.Type != "CAA" {
		panic("assertion failed: SetTargetCAA called when .Type is not CAA")
	}

	if tag != "issue" && tag != "issuewild" && tag != "iodef" {
		return fmt.Errorf("CAA tag (%v) is not one of issue/issuewild/iodef", tag)
	}

	return nil
}

// SetTargetCAAStrings is like SetTargetCAA but accepts strings.
func (rc *RecordConfig) SetTargetCAAStrings(flag, tag, target string) error {
	i64flag, err := strconv.ParseUint(flag, 10, 8)
	if err != nil {
		return fmt.Errorf("CAA flag does not fit in 8 bits: %w", err)
	}
	return rc.SetTargetCAA(uint8(i64flag), tag, target)
}

// SetTargetCAAString is like SetTargetCAA but accepts one big string.
// Ex: `0 issue "letsencrypt.org"`
func (rc *RecordConfig) SetTargetCAAString(s string) error {
	part := strings.Fields(s)
	if len(part) != 3 {
		return fmt.Errorf("CAA value does not contain 3 fields: (%#v)", s)
	}
	return rc.SetTargetCAAStrings(part[0], part[1], StripQuotes(part[2]))
}
