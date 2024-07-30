package rtypemx

import (
	"fmt"
	"strconv"
	"strings"
)

// SetTargetMX sets the MX fields.
func (rdat *MX) SetTargetMX(pref uint16, target string) error {
	rdat.Preference = pref
	rdat.Mx = target
	return nil
}

// SetTargetMXStrings is like SetTargetMX but accepts strings.
func (rdat *MX) SetTargetMXStrings(pref, target string) error {
	u64pref, err := strconv.ParseUint(pref, 10, 16)
	if err != nil {
		return fmt.Errorf("can't parse MX data: %w", err)
	}
	return rdat.SetTargetMX(uint16(u64pref), target)
}

// SetTargetMXString is like SetTargetMX but accepts one big string.
func (rdat *MX) SetTargetMXString(s string) error {
	part := strings.Fields(s)
	if len(part) != 2 {
		return fmt.Errorf("MX value does not contain 2 fields: (%#v)", s)
	}
	return rdat.SetTargetMXStrings(part[0], part[1])
}
