package models

// methods that make RecordConfig meet the dns.RR interface.

import (
	"fmt"
	"log"
	"strings"

	"github.com/miekg/dns"
)

//// Header Header returns the header of an resource record.
//func (rc *RecordConfig) Header() *dns.RR_Header {
//	log.Fatal("Header not implemented")
//	return nil
//}

// String returns the text representation of the resource record.
func (rc *RecordConfig) String() string {
	return rc.GetTargetCombined()
}

//// copy returns a copy of the RR
//func (rc *RecordConfig) copy() dns.RR {
//	log.Fatal("Copy not implemented")
//	return dns.TypeToRR[dns.TypeA]()
//}
//
//// len returns the length (in octets) of the uncompressed RR in wire format.
//func (rc *RecordConfig) len() int {
//	log.Fatal("len not implemented")
//	return 0
//}
//
//// pack packs an RR into wire format.
//func (rc *RecordConfig) pack([]byte, int, map[string]int, bool) (int, error) {
//	log.Fatal("pack not implemented")
//	return 0, nil
//}

// Conversions

// RRstoRCs converts []dns.RR to []RecordConfigs.
func RRstoRCs(rrs []dns.RR, origin string) Records {
	rcs := make(Records, 0, len(rrs))
	for _, r := range rrs {
		rc := RRtoRC(r, origin)
		rcs = append(rcs, &rc)
	}
	return rcs
}

// RRtoRC converts dns.RR to RecordConfig
func RRtoRC(rr dns.RR, origin string) RecordConfig {
	// Convert's dns.RR into our native data type (RecordConfig).
	// Records are translated directly with no changes.
	header := rr.Header()
	rc := new(RecordConfig)
	rc.Type = dns.TypeToString[header.Rrtype]
	rc.TTL = header.Ttl
	rc.Original = rr
	rc.SetLabelFromFQDN(strings.TrimSuffix(header.Name, "."), origin)
	switch v := rr.(type) { // #rtype_variations
	case *dns.A:
		panicInvalid(rc.SetTarget(v.A.String()))
	case *dns.AAAA:
		panicInvalid(rc.SetTarget(v.AAAA.String()))
	case *dns.CAA:
		panicInvalid(rc.SetTargetCAA(v.Flag, v.Tag, v.Value))
	case *dns.CNAME:
		panicInvalid(rc.SetTarget(v.Target))
	case *dns.DS:
		panicInvalid(rc.SetTargetDS(v.KeyTag, v.Algorithm, v.DigestType, v.Digest))
	case *dns.MX:
		panicInvalid(rc.SetTargetMX(v.Preference, v.Mx))
	case *dns.NS:
		panicInvalid(rc.SetTarget(v.Ns))
	case *dns.PTR:
		panicInvalid(rc.SetTarget(v.Ptr))
	case *dns.NAPTR:
		panicInvalid(rc.SetTargetNAPTR(v.Order, v.Preference, v.Flags, v.Service, v.Regexp, v.Replacement))
	case *dns.SOA:
		panicInvalid(rc.SetTargetSOA(v.Ns, v.Mbox, v.Serial, v.Refresh, v.Retry, v.Expire, v.Minttl))
	case *dns.SRV:
		panicInvalid(rc.SetTargetSRV(v.Priority, v.Weight, v.Port, v.Target))
	case *dns.SSHFP:
		panicInvalid(rc.SetTargetSSHFP(v.Algorithm, v.Type, v.FingerPrint))
	case *dns.TLSA:
		panicInvalid(rc.SetTargetTLSA(v.Usage, v.Selector, v.MatchingType, v.Certificate))
	case *dns.TXT:
		panicInvalid(rc.SetTargetTXTs(v.Txt))
	default:
		log.Fatalf("rrToRecord: Unimplemented zone record type=%s (%v)\n", rc.Type, rr)
	}
	return *rc
}

func panicInvalid(err error) {
	if err != nil {
		panic(fmt.Errorf("unparsable record received from BIND: %w", err))
	}
}
