package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// BUNNY_DNS_PZ

func init() {
	Register(TypeBUNNY_DNS_PZ, "BUNNY_DNS_PZ", func() dnsv2.RR { return new(BUNNY_DNS_PZ) }, privatetypesrdata.MakeBUNNY_DNS_PZ)
}

const TypeBUNNY_DNS_PZ = 65313

type BUNNY_DNS_PZ struct {
	Hdr dnsv2.Header
}

// Typer interface.
func (rr *BUNNY_DNS_PZ) Type() uint16 { return TypeBUNNY_DNS_PZ }

// RR interface.
func (rr *BUNNY_DNS_PZ) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *BUNNY_DNS_PZ) Len() int              { return rr.Hdr.Len() }
func (rr *BUNNY_DNS_PZ) Data() dnsv2.RDATA {
	return &privatetypesrdata.BUNNY_DNS_PZ{}
}
func (rr *BUNNY_DNS_PZ) Clone() dnsv2.RR {
	return &BUNNY_DNS_PZ{rr.Hdr}
}
func (rr *BUNNY_DNS_PZ) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tBUNNY_DNS_PZ"
}

// Parser interface.
func (rr *BUNNY_DNS_PZ) Parse(tokens []string, _ string) error {
	return nil
}
