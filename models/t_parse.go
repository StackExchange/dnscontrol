package models

import (
	"net"

	"github.com/pkg/errors"
)

// PopulateFromString populates a RecordConfig given a type and string.
// Many providers give all the parameters of a resource record in one big
// string (all the parameters of an MX, SRV, CAA, etc). Rather than have
// each provider rewrite this code many times, here's a helper function to use.
//
// At this time, the idiom is to panic rather than continue with potentially
// misunderstood data. We do this panic() at the provider level.
// Therefore the typical calling sequence is:
//     if err := rc.PopulateFromString(rtype, value, origin); err != nil {
//         panic(errors.Wrap(err, "unparsable record received from provider"))
//     }
func (r *RecordConfig) PopulateFromString(rtype, contents, origin string) error {
	if r.Type != "" && r.Type != rtype {
		panic(errors.Errorf("assertion failed: rtype already set (%s) (%s)", rtype, r.Type))
	}
	switch r.Type = rtype; rtype { // #rtype_variations
	case "A":
		ip := net.ParseIP(contents)
		if ip == nil || ip.To4() == nil {
			return errors.Errorf("A record with invalid IP: %s", contents)
		}
		return r.SetTargetIP(ip) // Reformat to canonical form.
	case "AAAA":
		ip := net.ParseIP(contents)
		if ip == nil || ip.To16() == nil {
			return errors.Errorf("AAAA record with invalid IP: %s", contents)
		}
		return r.SetTargetIP(ip) // Reformat to canonical form.
	case "ANAME", "CNAME", "NS", "PTR":
		return r.SetTarget(contents)
	case "CAA":
		return r.SetTargetCAAString(contents)
	case "MX":
		return r.SetTargetMXString(contents)
	case "SRV":
		return r.SetTargetSRVString(contents)
	case "TLSA":
		return r.SetTargetTLSAString(contents)
	case "TXT":
		return r.SetTargetTXTString(contents)
	default:
		return errors.Errorf("Unknown rtype (%s) when parsing (%s) domain=(%s)",
			rtype, contents, origin)
	}
}
