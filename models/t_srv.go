package models

import (
	"fmt"
	"strconv"
	"strings"
)

// SetTargetSRVStrings is like SetTargetSRV but accepts all parameters as strings.
func (rc *RecordConfig) SetTargetSRVStrings(priority, weight, port, target string) (err error) {
	rc.Type = "SRV"

	rdata, err := ParseSRV([]string{priority, weight, port, target}, "")
	if err != nil {
		return err
	}
	return RecordUpdateFields(rc, rdata, nil)
}

// SetTargetSRVPriorityString is like SetTargetSRV but accepts priority as an
// uint16 and the rest of the values joined in a string that needs to be parsed.
// This is a helper function that comes in handy when a provider re-uses the MX preference
// field as the SRV priority.
func (rc *RecordConfig) SetTargetSRVPriorityString(priority uint16, s string) error {
	var rdata SRV
	var err error

	part := strings.Fields(s)
	switch len(part) {
	case 3:
		rdata, err = ParseSRV([]string{strconv.Itoa(int(priority)), part[0], part[1], part[2]}, "")
	case 2:
		rdata, err = ParseSRV([]string{strconv.Itoa(int(priority)), part[0], part[1], "."}, "")
	default:
		return fmt.Errorf("SRV value does not contain 3 fields: (%#v)", s)
	}
	if err != nil {
		return err
	}
	return RecordUpdateFields(rc, rdata, nil)
}

// SetTargetSRVString is like SetTargetSRV but accepts one big string to be parsed.
func (rc *RecordConfig) SetTargetSRVString(s string) error {
	part := strings.Fields(s)
	return rc.SetTargetSRVStrings(part[0], part[1], part[2], part[3])
}
