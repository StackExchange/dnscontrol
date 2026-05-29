package models

import (
	"bytes"
	"fmt"
	"strings"

	dnsv2 "codeberg.org/miekg/dns"
	dnsrdatav2 "codeberg.org/miekg/dns/rdata"
	"codeberg.org/miekg/dns/svcb"
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

	if rc.SvcPriority == 0 {
		rc.RDATA = dnsrdatav2.SVCB{Priority: rc.SvcPriority, Target: rc.GetTargetField()}
	} else {
		rd, err := dnsv2.NewData(dnsv2.TypeSVCB, fmt.Sprintf("%d %s %s", rc.SvcPriority, rc.GetTargetField(), rc.SvcParams), ".")
		if err != nil {
			panic(fmt.Sprintf("BUG: Failed to create RDATA for HTTPS record: %v", err))
		}
		rc.RDATA = rd
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

	if rc.SvcPriority == 0 {
		rc.RDATA = dnsrdatav2.SVCB{Priority: rc.SvcPriority, Target: rc.GetTargetField()}
	} else {
		rd, err := dnsv2.NewData(dnsv2.TypeSVCB, fmt.Sprintf("%d %s %s", rc.SvcPriority, rc.GetTargetField(), rc.SvcParams), origin)
		if err != nil {
			panic(fmt.Sprintf("BUG: Failed to create RDATA for HTTPS record: %v", err))
		}
		rc.RDATA = rd
	}
	rc.FixUp(".")

	return nil
}

// func convertSVCBv1v2(params []dnsv1.SVCBKeyValue) ([]svcbv2.Pair, error) {
// 	var value []svcbv2.Pair
// 	for _, kvV1 := range params {
// 		kV1 := kvV1.Key().String()
// 		keyCodeV2 := svcbv2.StringToKey(kV1)
// 		vV1 := kvV1.String()
// 		if len(vV1) > 2 && vV1[0] == '"' && vV1[len(vV1)-1] == '"' {
// 			panic("V has quotes")
// 		}
// 		fmt.Printf("DEBUG: convertSVCBv1v2: k=%s keyCode=%d v1=%s\n", kV1, keyCodeV2, vV1)

// 		pairFn := svcbv2.KeyToPair(keyCodeV2)
// 		if pairFn == nil {
// 			return nil, fmt.Errorf("failed to lookup svc key: %s", kV1)
// 		}
// 		pair := pairFn()
// 		if svcbv2.PairToKey(pair) != keyCodeV2 {
// 			return nil, fmt.Errorf("key constant is not in sync: %v", keyCodeV2)
// 		}
// 		err := svcbv2.Parse(pair, vV1, "")
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to parse svc pair: %s", kV1)
// 		}

// 		vV2 := pair.String()
// 		if len(vV2) > 2 && vV2[0] == '"' && vV2[len(vV2)-1] == '"' {
// 			panic("V2 has quotes")
// 		}
// 		if vV1 != vV2 {
// 			panic(fmt.Sprintf("conversion from v1 to v2 is not stable: key=%s v1=%s v2=%s", kV1, vV1, vV2))
// 		}

// 		value = append(value, pair)
// 	}

// 	return value, nil
// }

func ModifySVCBForComparison(existing, desired Records) Records {

	// Clone the "desired" list. Its an array of pointers, so we clone the pointers. We can replace any record we want in the "desired" list without mutating the original.
	newDesired := make(Records, len(desired))
	copy(newDesired, desired)

	// Build the list of existing ECH values.
	echs := gatherEchValues(existing)

	// Scan desired for ech=IGNORE.  Replace any records.
	recs, edits := replaceSvcbIgnores(newDesired, echs)

	if edits {
		return recs
	}
	// No changes were made, so we can return the original "desired" list to save memory.
	return desired
}

// gatherEchValues builds a map of FQDN to ECH values for all SVCB and HTTPS
// records in the given set of records.  This is used to support the
// "ech=IGNORE" feature, where we want to ignore changes in the ECH value when
// comparing records, but still show the ECH value in the output for debugging
// purposes.
func gatherEchValues(recs Records) map[string]*svcbv2.ECHCONFIG {
	echs := map[string]*svcbv2.ECHCONFIG{}
	for _, rec := range recs {
		if rec.TypeNum == dnsv2.TypeSVCB || rec.TypeNum == dnsv2.TypeHTTPS {
			if value, ok := rec.GetSVCBEchConfig(); ok {
				echs[rec.NameFQDN] = value
			}
		}
	}
	return echs
}

// GetSVCBEchConfig returns the value of the ECH parameter. The value is a pointer to a clone.
func (rc *RecordConfig) GetSVCBEchConfig() (*svcbv2.ECHCONFIG, bool) {
	if rc.TypeNum != dnsv2.TypeSVCB && rc.TypeNum != dnsv2.TypeHTTPS {
		panic("assertion failed: GetSVCBParam called when .Type is not SVCB or HTTPS")
	}
	if rc.RDATA == nil {
		panic("assertion failed: SVCB/HTTPS record does not have RDATA set")
	}

	for _, param := range rc.RDATA.(*dnsrdatav2.SVCB).Value {
		key := svcbv2.PairToKey(param)
		if key == svcbv2.KeyEchConfig {
			p := param.(*svcbv2.ECHCONFIG)
			c := p.Clone()
			return c.(*svcbv2.ECHCONFIG), true
		}
	}
	return nil, false
}

func replaceSvcbIgnores(records Records, echs map[string]*svcbv2.ECHCONFIG) (Records, bool) {
	edits := false

	for i, rec := range records {
		// Skip rtypes we're not concerned with.
		if rec.TypeNum != dnsv2.TypeSVCB && rec.TypeNum != dnsv2.TypeHTTPS {
			continue
		}

		// Skip records that don't have ech=IGNORE.
		if ec, ok := rec.GetSVCBEchConfig(); ok && !bytes.Equal(ec.ECH, []byte("IGNORE")) {
			continue
		}

		// Now we know this record has ech=IGNORE.
		nRec, err := rec.Copy()
		if err != nil {
			panic(fmt.Sprintf("failed to copy record: %v", err))
		}
		if nEch, ok := echs[rec.NameFQDN]; ok {
			// This record has an ECH value, so we replace "ech=IGNORE" with the actual value in the comparables.
			nRec.RDATA = SVCBReplaceEch(nRec.RDATA.(*dnsrdatav2.SVCB), nEch)
		} else {
			// This record doesn't have an ECH value, so we delete "ech=IGNORE" from the comparables.
			nRec.RDATA = SVCBDeleteEch(nRec.RDATA.(*dnsrdatav2.SVCB))
		}

		// Clear the ComparableV3 so it will be regenerated with the new RDATA.
		nRec.ComparableV3 = ""
		nRec.FixUp(".")

		// Replace the record.
		records[i] = nRec
		edits = true
	}

	return records, edits
}

func SVCBReplaceEch(rr *dnsrdatav2.SVCB, echConfig *svcbv2.ECHCONFIG) *dnsrdatav2.SVCB {
	// This is a bit of a hack, but dnsrdatav2.SVCB doesn't have a Clone method.
	// We need to clone it to avoid mutating the original.
	// We replace "ech=" as we clone it.
	return &dnsrdatav2.SVCB{
		Priority: rr.Priority,
		Target:   rr.Target,
		Value: func() []svcb.Pair {
			pairs := make([]svcb.Pair, len(rr.Value))
			found := false
			for i, p := range rr.Value {
				if svcbv2.PairToKey(p) == svcbv2.KeyEchConfig {
					pairs[i] = echConfig
					found = true
				} else {
					pairs[i] = p.Clone()
				}
			}
			if !found {
				pairs = append(pairs, echConfig)
			}
			return pairs
		}(),
	}
}

func SVCBDeleteEch(rr *dnsrdatav2.SVCB) *dnsrdatav2.SVCB {
	// This is a bit of a hack, but dnsrdatav2.SVCB doesn't have a Clone method.
	// We need to clone it to avoid mutating the original.
	// We deelte "ech=" as we clone it.
	return &dnsrdatav2.SVCB{
		Priority: rr.Priority,
		Target:   rr.Target,
		Value: func() []svcb.Pair {
			pairs := make([]svcb.Pair, len(rr.Value))
			for i, p := range rr.Value {
				if svcbv2.PairToKey(p) != svcbv2.KeyEchConfig {
					pairs[i] = p.Clone()
				}
			}
			return pairs
		}(),
	}
}
