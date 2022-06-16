package models

import (
	"fmt"
	"net"
)

// PopulateFromString populates a RecordConfig given a type and string.
// Many providers give all the parameters of a resource record in one big
// string (all the parameters of an MX, SRV, CAA, etc). Rather than have
// each provider rewrite this code many times, here's a helper function to use.
//
// If this doesn't work for all rtypes, process the special cases then
// call this for the remainder.
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
	case "MX":
		return rc.SetTargetMXString(contents)
	case "NAPTR":
		return rc.SetTargetNAPTRString(contents)
	case "SOA":
		return rc.SetTargetSOAString(contents)
	case "SPF", "TXT":
		fmt.Printf("DEBUG: popFrmStr txt=%q\n", contents)
		return rc.SetTargetTXTString(contents)
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
