package models

// methods that make RecordConfig meet the dns.RR interface.

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

// String returns the text representation of the resource record.
func (rc *RecordConfig) String() string {
	return rc.GetTargetCombined()
}

// Conversions

// RRstoRCs converts []dns.RR to []RecordConfigs.
func RRstoRCs(rrs []dns.RR, origin string) (Records, error) {
	rcs := make(Records, 0, len(rrs))
	for _, r := range rrs {
		rc, err := RRtoRC(r, origin)
		if err != nil {
			return nil, err
		}

		rcs = append(rcs, &rc)
	}
	return rcs, nil
}

// RRtoRC converts dns.RR to RecordConfig
func RRtoRC(rr dns.RR, origin string) (RecordConfig, error) {
	// Convert's dns.RR into our native data type (RecordConfig).
	// Records are translated directly with no changes.
	header := rr.Header()
	rc := new(RecordConfig)
	rc.Type = dns.TypeToString[header.Rrtype]
	rc.TTL = header.Ttl
	rc.Original = rr
	rc.SetLabelFromFQDN(strings.TrimSuffix(header.Name, "."), origin)
	var err error
	switch v := rr.(type) { // #rtype_variations
	case *dns.A:
		err = rc.SetTarget(v.A.String())
	case *dns.AAAA:
		err = rc.SetTarget(v.AAAA.String())
	case *dns.CAA:
		err = rc.SetTargetCAA(v.Flag, v.Tag, v.Value)
	case *dns.CNAME:
		err = rc.SetTarget(v.Target)
	case *dns.DHCID:
		err = rc.SetTarget(v.Digest)
	case *dns.DS:
		err = rc.SetTargetDS(v.KeyTag, v.Algorithm, v.DigestType, v.Digest)
	case *dns.LOC:
		err = rc.SetTargetLOC(v.Version, v.Latitude, v.Longitude, v.Altitude, v.Size, v.HorizPre, v.VertPre)
	case *dns.MX:
		err = rc.SetTargetMX(v.Preference, v.Mx)
	case *dns.NAPTR:
		err = rc.SetTargetNAPTR(v.Order, v.Preference, v.Flags, v.Service, v.Regexp, v.Replacement)
	case *dns.NS:
		err = rc.SetTarget(v.Ns)
	case *dns.PTR:
		err = rc.SetTarget(v.Ptr)
	case *dns.SOA:
		err = rc.SetTargetSOA(v.Ns, v.Mbox, v.Serial, v.Refresh, v.Retry, v.Expire, v.Minttl)
	case *dns.SRV:
		err = rc.SetTargetSRV(v.Priority, v.Weight, v.Port, v.Target)
	case *dns.SSHFP:
		err = rc.SetTargetSSHFP(v.Algorithm, v.Type, v.FingerPrint)
	case *dns.TLSA:
		err = rc.SetTargetTLSA(v.Usage, v.Selector, v.MatchingType, v.Certificate)
	case *dns.TXT:
		err = rc.SetTargetTXTs(v.Txt)
	default:
		return *rc, fmt.Errorf("rrToRecord: Unimplemented zone record type=%s (%v)", rc.Type, rr)
	}
	if err != nil {
		return *rc, fmt.Errorf("unparsable record received: %w", err)
	}
	return *rc, nil
}
