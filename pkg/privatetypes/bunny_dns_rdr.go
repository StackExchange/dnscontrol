package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// BUNNY_DNS_RDR

func init() {
	Register(TypeBUNNY_DNS_RDR, "BUNNY_DNS_RDR", func() dnsv2.RR { return new(BUNNY_DNS_RDR) }, privatetypesrdata.MakeBUNNY_DNS_RDR)
}

const TypeBUNNY_DNS_RDR = 65320

type BUNNY_DNS_RDR struct {
	Hdr dnsv2.Header
}

// Typer interface.
func (rr *BUNNY_DNS_RDR) Type() uint16 { return TypeBUNNY_DNS_RDR }

// RR interface.
func (rr *BUNNY_DNS_RDR) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *BUNNY_DNS_RDR) Len() int              { return rr.Hdr.Len() }
func (rr *BUNNY_DNS_RDR) Data() dnsv2.RDATA {
	return &privatetypesrdata.BUNNY_DNS_RDR{}
}
func (rr *BUNNY_DNS_RDR) Clone() dnsv2.RR {
	return &BUNNY_DNS_RDR{rr.Hdr}
}
func (rr *BUNNY_DNS_RDR) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tBUNNY_DNS_RDR\t" + rr.Data().String()
}

// Parser interface.
func (rr *BUNNY_DNS_RDR) Parse(tokens []string, _ string) error {
	return nil
}
