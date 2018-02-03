package models

import (
	"net"

	"github.com/pkg/errors"
)

// PopulateFromString populates a RecordConfig given a type and string.
// Many providers give all the parameters of a resource record in one big
// string (all the parameters of an MX, SRV, CAA, etc). Rather than have
// each provider rewrite this code many times, here's a helper function to use.
// NOTE: This will panic if contents can not be parsed or is invalid.
func (r *RecordConfig) PopulateFromString(rtype, contents, origin string) error {
	if r.Type != "" && r.Type != rtype {
		panic(errors.Errorf("rtype already set (%s) (%s)", rtype, r.Type))
	}
	r.Type = rtype
	switch rtype { // #rtype_variations
	case "A":
		ip := net.ParseIP(contents)
		if ip == nil || ip.To4() == nil {
			return errors.Errorf("A record with invalid IP: %s", contents)
		}
		r.SetTarget(ip.String()) // Reformat to canonical form.
	case "AAAA":
		ip := net.ParseIP(contents)
		if ip == nil || ip.To16() == nil {
			return errors.Errorf("AAAA record with invalid IP: %s", contents)
		}
		r.SetTarget(ip.String()) // Reformat to canonical form.
	case "ANAME", "CNAME", "NS", "PTR":
		r.SetTarget(contents)
	case "CAA":
		r.SetTargetCAAString(contents)
	case "MX":
		r.SetTargetMXString(contents)
	case "SRV":
		r.SetTargetSRVString(contents)
	case "TLSA":
		r.SetTargetTLSAString(contents)
	case "TXT":
		r.SetTargetTXTString(contents)
	default:
		return errors.Errorf("Unknown rtype (%s) when parsing (%s) domain=(%s)",
			rtype, contents, origin)
	}
	return nil
}
