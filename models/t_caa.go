package models

import (
	"fmt"
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/pkg/rfc1035"
)

// SetTargetCAAStrings is like SetTargetCAA but accepts strings.
func (rc *RecordConfig) SetTargetCAAStrings(flag, tag, target string) error {
	//fmt.Printf("DEBUG: CAA TESTs: %v %v %v\n", flag, tag, target)
	i64flag, err := strconv.ParseUint(flag, 10, 8)
	if err != nil {
		return fmt.Errorf("CAA flag does not fit in 8 bits: %w", err)
	}
	return rc.SetTargetCAA(uint8(i64flag), tag, target)
}

// SetTargetCAAString is like SetTargetCAA but accepts one big string.
// Ex: `0 issue "letsencrypt.org"`
func (rc *RecordConfig) SetTargetCAAString(s string) error {
	//fmt.Printf("DEBUG: CAA TEST: %q\n", s)
	//part, err := ParseQuotedFields(s)
	part, err := rfc1035.Fields(s)
	if err != nil {
		return err
	}
	if len(part) != 3 {
		return fmt.Errorf("CAA value does not contain 3 fields: (%#v)", s)
	}
	return rc.SetTargetCAAStrings(part[0], part[1], part[2])
}
