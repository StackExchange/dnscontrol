package models

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// SetTargetSRV sets the SRV fields.
func (rc *RecordConfig) SetTargetSRV(priority, weight, port uint16, target string) error {
	rc.SrvPriority = priority
	rc.SrvWeight = weight
	rc.SrvPort = port
	rc.Target = target
	if rc.Type == "" {
		rc.Type = "SRV"
	}
	if rc.Type != "SRV" {
		panic("assertion failed: SetTargetSRV called when .Type is not SRV")
	}
	return nil
}

// setTargetIntAndStrings is like SetTargetSRV but accepts priority as an int, the other parameters as strings.
func (rc *RecordConfig) setTargetIntAndStrings(priority uint16, weight, port, target string) (err error) {
	var i64weight, i64port uint64
	if i64weight, err = strconv.ParseUint(weight, 10, 16); err == nil {
		if i64port, err = strconv.ParseUint(port, 10, 16); err == nil {
			return rc.SetTargetSRV(priority, uint16(i64weight), uint16(i64port), target)
		}
	}
	return errors.Wrap(err, "SRV value too big for uint16")
}

// SetTargetSRVStrings is like SetTargetSRV but accepts all parameters as strings.
func (rc *RecordConfig) SetTargetSRVStrings(priority, weight, port, target string) (err error) {
	var i64priority uint64
	if i64priority, err = strconv.ParseUint(priority, 10, 16); err == nil {
		return rc.setTargetIntAndStrings(uint16(i64priority), weight, port, target)
	}
	return errors.Wrap(err, "SRV value too big for uint16")
}

// SetTargetSRVPriorityString is like SetTargetSRV but accepts priority as an
// uint16 and the rest of the values joined in a string that needs to be parsed.
// This is a helper function that comes in handy when a provider re-uses the MX preference
// field as the SRV priority.
func (rc *RecordConfig) SetTargetSRVPriorityString(priority uint16, s string) error {
	part := strings.Fields(s)
	if len(part) != 3 {
		return errors.Errorf("SRV value does not contain 3 fields: (%#v)", s)
	}
	return rc.setTargetIntAndStrings(priority, part[0], part[1], part[2])
}

// SetTargetSRVString is like SetTargetSRV but accepts one big string to be parsed.
func (rc *RecordConfig) SetTargetSRVString(s string) error {
	part := strings.Fields(s)
	if len(part) != 4 {
		return errors.Errorf("SRC value does not contain 4 fields: (%#v)", s)
	}
	return rc.SetTargetSRVStrings(part[0], part[1], part[2], part[3])
}
