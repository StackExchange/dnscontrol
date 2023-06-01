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
// function for everything else.
//
//		  var err error
//		  switch rType {
//		  case "MX":
//	       // MX priority in a separate field.
//	       if err := rc.SetTargetMX(cr.Priority, target); err != nil {
//	         return nil, fmt.Errorf("unparsable MX record received from cloudflare: %w", err)
//	       }
//		  case "TXT":
//	       // TXT records are stored verbatim; no quoting/escaping to parse.
//			  err = rc.SetTargetTXT(target)
//	       // ProTip: Use rc.SetTargetTXTs(manystrings) if the API or parser returns a list of substrings.
//		  default:
//			  err = rec.PopulateFromString(rType, target, origin)
//		  }
//		  if err != nil {
//			  return nil, fmt.Errorf("unparsable record received from CHANGE_TO_PROVDER_NAME: %w", err)
//		  }
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
		return rc.SetTargetTXT(contents)
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
