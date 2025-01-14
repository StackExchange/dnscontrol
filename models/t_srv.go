package models

import (
	"fmt"
	"strconv"
	"strings"
)

// SetTargetSRV sets the SRV fields.
func (rc *RecordConfig) SetTargetSRV(priority, weight, port uint16, target string) error {
	return rc.PopulateFieldsSRV(priority, weight, port, target, nil)
}

// SetTargetSRVStrings is like SetTargetSRV but accepts all parameters as strings.
func (rc *RecordConfig) SetTargetSRVStrings(priority, weight, port, target string) (err error) {
	return PopulateFromRawSRV(rc, []string{rc.Name, priority, weight, port, target}, nil, "")
}

// SetTargetSRVPriorityString is like SetTargetSRV but accepts priority as an
// uint16 and the rest of the values joined in a string that needs to be parsed.
// This is a helper function that comes in handy when a provider re-uses the MX preference
// field as the SRV priority.
func (rc *RecordConfig) SetTargetSRVPriorityString(priority uint16, s string) error {
	part := strings.Fields(s)
	switch len(part) {
	case 3:
		return PopulateFromRawSRV(rc, []string{rc.Name, strconv.Itoa(int(priority)), part[0], part[1], part[2]}, nil, "")
	case 2:
		return PopulateFromRawSRV(rc, []string{rc.Name, strconv.Itoa(int(priority)), part[0], part[1], "."}, nil, "")
	default:
		return fmt.Errorf("SRV value does not contain 3 fields: (%#v)", s)
	}
}

// SetTargetSRVString is like SetTargetSRV but accepts one big string to be parsed.
func (rc *RecordConfig) SetTargetSRVString(s string) error {
	part := strings.Fields(s)
	if len(part) != 4 {
		return fmt.Errorf("SRV value does not contain 4 fields: (%#v)", s)
	}
	return rc.SetTargetSRVStrings(part[0], part[1], part[2], part[3])
}
