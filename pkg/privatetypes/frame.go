package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// FRAME

func init() {
	Register(TypeFRAME, "FRAME", func() dnsv2.RR { return new(FRAME) }, privatetypesrdata.MakeFRAME)
}

const TypeFRAME = 65312

type FRAME struct {
	Hdr    dnsv2.Header
	Target string
}

// Typer interface.
func (rr *FRAME) Type() uint16 { return TypeFRAME }

// RR interface.
func (rr *FRAME) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *FRAME) Len() int              { return rr.Hdr.Len() }
func (rr *FRAME) Data() dnsv2.RDATA {
	return &privatetypesrdata.FRAME{}
}
func (rr *FRAME) Clone() dnsv2.RR {
	return &FRAME{rr.Hdr, rr.Target}
}
func (rr *FRAME) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tFRAME\t" + rr.Data().String()
}

// Parser interface.
func (rr *FRAME) Parse(tokens []string, _ string) error {
	return nil
}
