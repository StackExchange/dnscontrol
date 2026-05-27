package models

import (
	"fmt"
	"strings"

	dnsv2 "codeberg.org/miekg/dns"
	dnsrdatav2 "codeberg.org/miekg/dns/rdata"
	svcbv2 "codeberg.org/miekg/dns/svcb"
	dnsv1 "github.com/miekg/dns"
)

func (rc *RecordConfig) targetCombinedSVCBRaw() string {
	if rc.SvcParams == "" {
		return fmt.Sprintf("%d %s", rc.SvcPriority, rc.target)
	}
	return fmt.Sprintf("%d %s %s", rc.SvcPriority, rc.target, rc.SvcParams)
}

// SetTargetSVCB sets the SVCB fields.
func (rc *RecordConfig) SetTargetSVCB(priority uint16, target string, params []dnsv1.SVCBKeyValue) error {
	rc.SvcPriority = priority
	if err := rc.SetTarget(target); err != nil {
		return err
	}
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

	// Hack to set .RDATA without importing miekg/dns in pkg/rtypecontrol/fixlegacy.go
	valuev2, err := convertSVCBv1v2(params)
	if err != nil {
		return fmt.Errorf("failed to convert SVCB parameters from v1 to v2: %w", err)
	}
	rc.RDATA = dnsrdatav2.SVCB{Priority: rc.SvcPriority, Target: target, Value: valuev2}
	rc.ComparableV3 = rc.RDATA.String()

	return nil
}

// SetTargetSVCBString is like SetTargetSVCB but accepts one big string and the origin so parsing can be done using miekg/dns.
func (rc *RecordConfig) SetTargetSVCBString(origin, contents string) error {
	if rc.Type == "" {
		rc.Type = "SVCB"
	}
	record, err := dnsv1.NewRR(fmt.Sprintf("%s. %s %s", origin, rc.Type, contents))
	if err != nil {
		return fmt.Errorf("could not parse SVCB record: %w", err)
	}

	// Hack to set .RDATA without importing miekg/dns in pkg/rtypecontrol/fixlegacy.go
	var rty uint16
	switch record.(type) {
	case *dnsv1.HTTPS:
		rty = dnsv1.TypeHTTPS
	case *dnsv1.SVCB:
		rty = dnsv1.TypeSVCB
	default:
		return fmt.Errorf("unexpected record type after parsing SVCB record: %T", record)
	}
	rrv2, err := dnsv2.NewData(rty, contents, origin)
	if err != nil {
		return fmt.Errorf("could not parse SVCB record: %w", err)
	}
	rc.RDATA = rrv2

	switch r := record.(type) {
	case *dnsv1.HTTPS:
		return rc.SetTargetSVCB(r.Priority, r.Target, r.Value)
	case *dnsv1.SVCB:
		return rc.SetTargetSVCB(r.Priority, r.Target, r.Value)
	}
	return nil
}

func convertSVCBv1v2(params []dnsv1.SVCBKeyValue) ([]svcbv2.Pair, error) {
	var value []svcbv2.Pair
	for _, kv := range params {
		k := kv.Key().String()
		keyCode := svcbv2.StringToKey(k)
		v := kv.String()

		pairFn := svcbv2.KeyToPair(keyCode)
		if pairFn == nil {
			return nil, fmt.Errorf("failed to lookup svc key: %s", k)
		}
		pair := pairFn()
		if svcbv2.PairToKey(pair) != keyCode {
			return nil, fmt.Errorf("key constant is not in sync: %v", keyCode)
		}
		err := svcbv2.Parse(pair, v, "")
		if err != nil {
			return nil, fmt.Errorf("failed to parse svc pair: %s", k)
		}

		value = append(value, pair)
	}

	return value, nil
}
