package models

// methods that make RecordConfig meet the dns.RR interface.

import (
	"fmt"
	"log"
	"strings"

	"github.com/miekg/dns"
)

// Header Header returns the header of an resource record.
func (rc *RecordConfig) Header() *dns.RR_Header {
	log.Fatal("Header not implemented")
	return nil
}

// String returns the text representation of the resource record.
func (rc *RecordConfig) String() string {

	//  rdtype, ok := dns.StringToType[rc.Type]
	//  if !ok {
	//    log.Fatalf("No such DNS type as (%#v)\n", rc.Type)
	//  }

	return rc.GetTargetCombined()
}

// copy returns a copy of the RR
func (rc *RecordConfig) copy() dns.RR {
	log.Fatal("Copy not implemented")
	return dns.TypeToRR[dns.TypeA]()
}

// len returns the length (in octets) of the uncompressed RR in wire format.
func (rc *RecordConfig) len() int {
	log.Fatal("len not implemented")
	return 0
}

// pack packs an RR into wire format.
func (rc *RecordConfig) pack([]byte, int, map[string]int, bool) (int, error) {
	log.Fatal("pack not implemented")
	return 0, nil
}

// Conversions

// RRstoRCs converts []dns.RR to []RecordConfigs.
func RRstoRCs(rrs []dns.RR, origin string, replaceSerial uint32) Records {
	rcs := make(Records, 0, len(rrs))
	var x uint32
	for _, r := range rrs {
		var rc RecordConfig
		//fmt.Printf("CONVERT: %+v\n", r)
		rc, x = RRtoRC(r, origin, replaceSerial)
		replaceSerial = x
		rcs = append(rcs, &rc)
	}
	return rcs
}

// RRtoRC converts dns.RR to RecordConfig
func RRtoRC(rr dns.RR, origin string, replaceSerial uint32) (RecordConfig, uint32) {
	// Convert's dns.RR into our native data type (RecordConfig).
	// Records are translated directly with no changes.
	// If it is an SOA for the apex domain and
	// replaceSerial != 0, change the serial to replaceSerial.
	// WARNING(tlim): This assumes SOAs do not have serial=0.
	// If one is found, we replace it with serial=1.
	var oldSerial, newSerial uint32
	header := rr.Header()
	rc := new(RecordConfig)
	//	rc = &RecordConfig{
	//		Type:     dns.TypeToString[header.Rrtype],
	//		TTL:      header.Ttl,
	//		Original: rr,
	//	}
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
	case *dns.MX:
		panicInvalid(rc.SetTargetMX(v.Preference, v.Mx))
	case *dns.NS:
		panicInvalid(rc.SetTarget(v.Ns))
	case *dns.PTR:
		panicInvalid(rc.SetTarget(v.Ptr))
	case *dns.NAPTR:
		panicInvalid(rc.SetTargetNAPTR(v.Order, v.Preference, v.Flags, v.Service, v.Regexp, v.Replacement))
	case *dns.SOA:
		oldSerial = v.Serial
		if oldSerial == 0 {
			// For SOA records, we never return a 0 serial number.
			oldSerial = 1
		}
		newSerial = v.Serial
		//if (dnsutil.TrimDomainName(rc.Name, origin+".") == "@") && replaceSerial != 0 {
		if rc.GetLabel() == "@" && replaceSerial != 0 {
			newSerial = replaceSerial
		}
		panicInvalid(rc.SetTarget(
			fmt.Sprintf("%v %v %v %v %v %v %v",
				v.Ns, v.Mbox, newSerial, v.Refresh, v.Retry, v.Expire, v.Minttl),
		))
		// FIXME(tlim): SOA should be handled by splitting out the fields.
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
	return *rc, oldSerial
}

func panicInvalid(err error) {
	if err != nil {
		panic(fmt.Errorf("unparsable record received from BIND: %w", err))
	}
}
