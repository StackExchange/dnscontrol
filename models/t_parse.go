package models

import (
	"fmt"
	"net"
)

// PopulateFromString populates a RecordConfig given a type and string.  Many
// providers give all the parameters of a resource record in one big string.
// This helper function lets you not re-invent the wheel.
//
// NOTE: You almost always want to special-case TXT records. Every provider
// seems to quote them differently.
//
// Recommended calling convention:  Process the exceptions first, then use the
// PopulateFromString function for everything else.
//
//      rtype := FILL_IN_TYPE
//     	var err error
//      rc := &models.RecordConfig{Type: rtype}
//      rc.SetLabelFromFQDN(FILL_IN_NAME, origin)
//      rc.TTL = uint32(FILL_IN_TTL)
//      rc.Original = FILL_IN_ORIGINAL // The raw data received from provider (if needed later)
//     	switch rtype {
//     	case "MX":
//     		// MX priority in a separate field.
//     		err = rc.SetTargetMX(cr.Priority, target)
//     	case "TXT":
//     		// TXT records are stored verbatim; no quoting/escaping to parse.
//     		err = rc.SetTargetTXT(target)
//     	default:
//     		err = rc.PopulateFromString(rtype, target, origin)
//     	}
//     	if err != nil {
//     		return nil, fmt.Errorf("unparsable record type=%q received from PROVDER_NAME: %w", rtype, err)
//     	}
//     	return rc, nil

func (rc *RecordConfig) PopulateFromString(rtype, contents, origin string) error {
	if rc.Type != "" && rc.Type != rtype {
		panic(fmt.Errorf("assertion failed: rtype already set (%s) (%s)", rtype, rc.Type))
	}
	switch rc.Type = rtype; rtype { // #rtype_variations
	case "A":
		ip := net.ParseIP(contents)
		if ip == nil || ip.To4() == nil {
			return fmt.Errorf("invalid IP in A record: %s", contents)
		}
		return rc.SetTargetIP(ip) // Reformat to canonical form.
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
	case "LOC":
		return rc.SetTargetLOCString(origin, contents)
	case "MX":
		return rc.SetTargetMXString(contents)
	case "NAPTR":
		return rc.SetTargetNAPTRString(contents)
	case "SOA":
		return rc.SetTargetSOAString(contents)
	case "SPF", "TXT":
		return rc.SetTargetTXTs(ParseQuotedTxt(contents))
		// Some day we'll switch to this:
		//return rc.SetTargetTXT(contents)
	case "SRV":
		return rc.SetTargetSRVString(contents)
	case "SSHFP":
		return rc.SetTargetSSHFPString(contents)
	case "TLSA":
		return rc.SetTargetTLSAString(contents)
	default:
		return fmt.Errorf("unknown rtype (%s) when parsing (%s) domain=(%s)",
			rtype, contents, origin)
	}
}
