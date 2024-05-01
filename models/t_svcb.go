package models

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

// SetTargetSVCB sets the SVCB fields.
func (rc *RecordConfig) SetTargetSVCB(priority uint16, target string, params []dns.SVCBKeyValue) error {
	rc.SvcPriority = priority
	rc.SetTarget(target)
	paramsStr := []string{}
	for _, kv := range params {
		paramsStr = append(paramsStr, fmt.Sprintf("%s=%s", kv.Key(), kv.String()))
	}
	rc.SvcParams = strings.Join(paramsStr, " ")
	if rc.Type == "" {
		rc.Type = "SVCB"
	}
	if rc.Type != "SVCB" && rc.Type != "HTTPS" {
		panic("assertion failed: SetTargetSVCB called when .Type is not SVCB or HTTPS")
	}
	return nil
}

// SetTargetSVCBString is like SetTargetSVCB but accepts one big string and the origin so parsing can be done using miekg/dns.
func (rc *RecordConfig) SetTargetSVCBString(origin, contents string) error {
	if rc.Type == "" {
		rc.Type = "SVCB"
	}
	record, err := dns.NewRR(fmt.Sprintf("%s. %s %s", origin, rc.Type, contents))
	if err != nil {
		return fmt.Errorf("could not parse SVCB record: %s", err)
	}
	switch r := record.(type) {
	case *dns.HTTPS:
		return rc.SetTargetSVCB(r.Priority, r.Target, r.Value)
	case *dns.SVCB:
		return rc.SetTargetSVCB(r.Priority, r.Target, r.Value)
	}
	return nil
}
