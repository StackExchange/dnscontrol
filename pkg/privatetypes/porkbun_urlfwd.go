package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// PORKBUN_URLFWD

func init() {
	Register(TypePORKBUN_URLFWD, "PORKBUN_URLFWD", func() dnsv2.RR { return new(PORKBUN_URLFWD) }, privatetypesrdata.MakePORKBUN_URLFWD)
}

const TypePORKBUN_URLFWD = 65321

type PORKBUN_URLFWD struct {
	Hdr dnsv2.Header
}

// Typer interface.
func (rr *PORKBUN_URLFWD) Type() uint16 { return TypePORKBUN_URLFWD }

// RR interface.
func (rr *PORKBUN_URLFWD) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *PORKBUN_URLFWD) Len() int              { return rr.Hdr.Len() }
func (rr *PORKBUN_URLFWD) Data() dnsv2.RDATA {
	return &privatetypesrdata.PORKBUN_URLFWD{}
}
func (rr *PORKBUN_URLFWD) Clone() dnsv2.RR {
	return &PORKBUN_URLFWD{rr.Hdr}
}
func (rr *PORKBUN_URLFWD) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tPORKBUN_URLFWD\t" + rr.Data().String()
}

// Parser interface.
func (rr *PORKBUN_URLFWD) Parse(tokens []string, _ string) error {
	return nil
}
