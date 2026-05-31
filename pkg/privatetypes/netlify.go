package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// NETLIFY

func init() {
	Register(TypeNETLIFY, "NETLIFY", func() dnsv2.RR { return new(NETLIFY) }, privatetypesrdata.MakeNETLIFY)
}

const TypeNETLIFY = 65316

type NETLIFY struct {
	Hdr dnsv2.Header
}

// Typer interface.
func (rr *NETLIFY) Type() uint16 { return TypeNETLIFY }

// RR interface.
func (rr *NETLIFY) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *NETLIFY) Len() int              { return rr.Hdr.Len() }
func (rr *NETLIFY) Data() dnsv2.RDATA {
	return &privatetypesrdata.NETLIFY{}
}
func (rr *NETLIFY) Clone() dnsv2.RR {
	return &NETLIFY{rr.Hdr}
}
func (rr *NETLIFY) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tNETLIFY" // RDATA is empty.
}

// Parser interface.
func (rr *NETLIFY) Parse(tokens []string, _ string) error {
	return nil
}
