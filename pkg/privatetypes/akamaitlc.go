package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// AKAMAITLC

func init() {
	Register(TypeAKAMAITLC, "AKAMAITLC", func() dnsv2.RR { return new(AKAMAITLC) }, privatetypesrdata.MakeAKAMAITLC)
}

const TypeAKAMAITLC = 65319

type AKAMAITLC struct {
	Hdr dnsv2.Header
}

// Typer interface.
func (rr *AKAMAITLC) Type() uint16 { return TypeAKAMAITLC }

// RR interface.
func (rr *AKAMAITLC) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *AKAMAITLC) Len() int              { return rr.Hdr.Len() }
func (rr *AKAMAITLC) Data() dnsv2.RDATA {
	return &privatetypesrdata.AKAMAITLC{}
}
func (rr *AKAMAITLC) Clone() dnsv2.RR {
	return &AKAMAITLC{rr.Hdr}
}
func (rr *AKAMAITLC) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tAKAMAITLC\t" + rr.Data().String()
}

// Parser interface.
func (rr *AKAMAITLC) Parse(tokens []string, _ string) error {
	return nil
}
