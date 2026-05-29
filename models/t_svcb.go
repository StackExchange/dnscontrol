package models

import (
	"fmt"
	"strings"

	dnsv2 "codeberg.org/miekg/dns"
	svcbv2 "codeberg.org/miekg/dns/svcb"
	dnsv1 "github.com/miekg/dns"
)

// func (rc *RecordConfig) targetCombinedSVCBRaw() string {
// 	if rc.SvcParams == "" {
// 		return fmt.Sprintf("%d %s", rc.SvcPriority, rc.target)
// 	}
// 	return fmt.Sprintf("%d %s %s", rc.SvcPriority, rc.target, rc.SvcParams)
// }

// SetTargetSVCB sets the SVCB fields.
func (rc *RecordConfig) SetTargetSVCB(priority uint16, target string, params []dnsv1.SVCBKeyValue) error {
	rc.SvcPriority = priority
	if err := rc.SetTarget(target); err != nil {
		return err
	}
	paramsStr := []string{}
	for _, kv := range params {
		paramsStr = append(paramsStr, fmt.Sprintf("%s=%q", kv.Key(), kv.String()))
	}
	rc.SvcParams = strings.Join(paramsStr, " ")
	if rc.Type == "" {
		rc.Type = "SVCB"
	}
	if rc.Type != "SVCB" && rc.Type != "HTTPS" {
		panic("assertion failed: SetTargetSVCB called when .Type is not SVCB or HTTPS")
	}

	// Hack to set .RDATA without importing miekg/dns in pkg/rtypecontrol/fixlegacy.go
	// valuev2, err := convertSVCBv1v2(params)
	// if err != nil {
	// 	return fmt.Errorf("failed to convert SVCB parameters from v1 to v2: %w", err)
	// }
	// rc.RDATA = dnsrdatav2.SVCB{Priority: rc.SvcPriority, Target: target, Value: valuev2}
	// rc.ComparableV3 = rc.RDATA.String() + "Z"
	rc.FixUp(".")

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
	for _, kvV1 := range params {
		kV1 := kvV1.Key().String()
		keyCodeV2 := svcbv2.StringToKey(kV1)
		vV1 := kvV1.String()
		if len(vV1) > 2 && vV1[0] == '"' && vV1[len(vV1)-1] == '"' {
			panic("V has quotes")
		}
		fmt.Printf("DEBUG: convertSVCBv1v2: k=%s keyCode=%d v1=%s\n", kV1, keyCodeV2, vV1)

		pairFn := svcbv2.KeyToPair(keyCodeV2)
		if pairFn == nil {
			return nil, fmt.Errorf("failed to lookup svc key: %s", kV1)
		}
		pair := pairFn()
		if svcbv2.PairToKey(pair) != keyCodeV2 {
			return nil, fmt.Errorf("key constant is not in sync: %v", keyCodeV2)
		}
		err := svcbv2.Parse(pair, vV1, "")
		if err != nil {
			return nil, fmt.Errorf("failed to parse svc pair: %s", kV1)
		}

		vV2 := pair.String()
		if len(vV2) > 2 && vV2[0] == '"' && vV2[len(vV2)-1] == '"' {
			panic("V2 has quotes")
		}
		if vV1 != vV2 {
			panic(fmt.Sprintf("conversion from v1 to v2 is not stable: key=%s v1=%s v2=%s", kV1, vV1, vV2))
		}

		value = append(value, pair)
	}

	return value, nil
}
