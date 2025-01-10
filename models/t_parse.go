package models

import (
	"fmt"
	"net"
	"strings"
)

// PopulateFromStringFunc populates a RecordConfig by parsing a common RFC1035-like format.
//
//	rtype: the resource record type (rtype)
//	contents: a string that contains all parameters of the record's rdata (see below)
//	txtFn: If rtype == "TXT", this function is used pre-process the contents. nil will simply store the contents unchanged.
//
// The "contents" field is the format used in RFC1035 zonefiles. It is the text
// after the rtype in a line of a zonefile.
// For example, in the line: foo IN MX 10 mx.example.com.
// "contents" includes everything after the "MX" and any whitespace.
//
// Typical values for txtFn include:
//
//	nil:  no parsing required.
//	txtutil.ParseQuoted: Parse using @TomOnTime's interpretation of RFC1035.
//	txtutil.ParseCombined: Parse using @miekg's interpretation of RFC1035.
//
// Many providers deliver record data in this format or something close to it.
// This function is provided to reduce the amount of duplicate code across
// providers.  If a particular rtype is not handled as a particular provider
// expects, simply handle it beforehand as a special case.
//
// Example 1: Normal use.
//
//	rtype := FILL_IN_RTYPE
//	rc := &models.RecordConfig{Type: rtype, TTL: FILL_IN_TTL}
//	rc.SetLabelFromFQDN(FILL_IN_NAME, origin)
//	rc.Original = FILL_IN_ORIGINAL // The raw data received from provider (if needed later)
//	if err = rc.PopulateFromStringFunc(rtype, target, origin, nil); err != nil {
//		return nil, fmt.Errorf("unparsable record type=%q received from PROVDER_NAME: %w", rtype, err)
//	}
//	return rc, nil
//
// Example 2: Use your own MX parser.
//
//	rtype := FILL_IN_RTYPE
//	rc := &models.RecordConfig{Type: rtype, TTL: FILL_IN_TTL}
//	rc.SetLabelFromFQDN(FILL_IN_NAME, origin)
//	rc.Original = FILL_IN_ORIGINAL // The raw data received from provider (if needed later)
//	switch rtype {
//	case "MX":
//		// MX priority in a separate field.
//		err = rc.SetTargetMX(cr.Priority, target)
//	default:
//		err = rc.PopulateFromString(rtype, target, origin)
//	}
//	if err != nil {
//		return nil, fmt.Errorf("unparsable record type=%q received from PROVDER_NAME: %w", rtype, err)
//	}
//	return rc, nil
func (rc *RecordConfig) PopulateFromStringFunc(rtype, contents, origin string, txtFn func(s string) (string, error)) error {
	if rc.Type != "" && rc.Type != rtype {
		return fmt.Errorf("assertion failed: rtype already set (%s) (%s)", rtype, rc.Type)
	}

	switch rc.Type = rtype; rtype { // #rtype_variations
	case "A":
		return PopulateARaw(rc, []string{rc.Name, contents}, nil, origin)
	case "AAAA":
		ip := net.ParseIP(contents)
		if ip == nil || ip.To16() == nil {
			return fmt.Errorf("invalid IP in AAAA record: %s", contents)
		}
		return rc.SetTargetIP(ip) // Reformat to canonical form.
	case "AKAMAICDN", "ALIAS", "ANAME", "CNAME", "NS", "PTR":
		return rc.SetTarget(contents)
	case "CAA":
		return rc.SetTargetCAAString(contents)
	case "DS":
		return rc.SetTargetDSString(contents)
	case "DNSKEY":
		return rc.SetTargetDNSKEYString(contents)
	case "DHCID":
		return rc.SetTarget(contents)
	case "DNAME":
		return rc.SetTarget(contents)
	case "LOC":
		return rc.SetTargetLOCString(origin, contents)
	case "MX":
		//fmt.Printf("DEBUG: contents=%q\n", contents)
		//fmt.Printf("DEBUG: PopulateMXRaw(rc, fields=%v, nil, %q)\n", append([]string{rc.Name}, strings.Fields(contents)...), origin)
		return PopulateFromRawMX(rc, append([]string{rc.Name}, strings.Fields(contents)...), nil, origin)
	case "NAPTR":
		return rc.SetTargetNAPTRString(contents)
	case "SOA":
		return rc.SetTargetSOAString(contents)
	case "SPF", "TXT":
		if txtFn == nil {
			return rc.SetTargetTXT(contents)
		}
		t, err := txtFn(contents)
		if err != nil {
			return fmt.Errorf("invalid TXT record: %s", contents)
		}
		return rc.SetTargetTXT(t)
	case "SRV":
		return PopulateSRVRaw(rc, append([]string{rc.Name}, strings.Fields(contents)...), nil, origin)
	case "SSHFP":
		return rc.SetTargetSSHFPString(contents)
	case "SVCB", "HTTPS":
		return rc.SetTargetSVCBString(origin, contents)
	case "TLSA":
		return rc.SetTargetTLSAString(contents)
	default:
		// return fmt.Errorf("unknown rtype (%s) when parsing (%s) domain=(%s)", rtype, contents, origin)
		return MakeUnknown(rc, rtype, contents, origin)
	}
}

// PopulateFromString populates a RecordConfig given a type and string.  It is equivalent to
// rc.PopulateFromStringFunc(rtype, contents, origin, nil)
func (rc *RecordConfig) PopulateFromString(rtype, contents, origin string) error {
	return rc.PopulateFromStringFunc(rtype, contents, origin, nil)
}
