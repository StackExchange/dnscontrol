package models

import (
	"fmt"
	"strings"
)

// SetTargetSRV sets the SRV fields.
func (rc *RecordConfig) SetTargetSRV(priority, weight, port uint16, target string) {
	rc.SrvPriority = priority
	rc.SrvWeight = weight
	rc.SrvPort = port
	rc.Target = target
	if rc.Type == "" {
		rc.Type = "SRV"
	}
	if rc.Type != "SRV" {
		panic("SetTargetSRV called when .Type is not SRV")
	}
}

// SetTargetSRVStrings is like SetTargetSRV but accepts strings.
func (rc *RecordConfig) SetTargetSRVStrings(priority, weight, port, target string) {
	rc.SetTargetSRV(atou16(priority), atou16(weight), atou16(port), target)
}

// SetTargetSRVString is like SetTargetSRV but accepts one big string.
func (rc *RecordConfig) SetTargetSRVString(s string) {
	part := strings.Fields(s)
	if len(part) != 4 {
		panic(fmt.Errorf("SRV value %#v contains too many fields", s))
	}
	rc.SetTargetSRVStrings(part[0], part[1], part[2], part[3])
}
