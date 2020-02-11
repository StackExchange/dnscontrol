package models

// methods that make RecordConfig meet the dns.RR interface.

import (
	"log"

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

	log.Fatal("String not implemented")
	return ""
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
