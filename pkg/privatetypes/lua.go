package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// LUA

func init() {
	Register(TypeLUA, "LUA", func() dnsv2.RR { return new(LUA) })
}

const TypeLUA = 65314

type LUA struct {
	Hdr dnsv2.Header
}

// Typer interface.
func (rr *LUA) Type() uint16 { return TypeLUA }

// RR interface.
func (rr *LUA) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *LUA) Len() int              { return rr.Hdr.Len() }
func (rr *LUA) Data() dnsv2.RDATA {
	return &privatetypesrdata.LUA{}
}
func (rr *LUA) Clone() dnsv2.RR {
	return &LUA{rr.Hdr}
}
func (rr *LUA) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tLUA"
}

// Parser interface.
func (rr *LUA) Parse(tokens []string, _ string) error {
	return nil
}
