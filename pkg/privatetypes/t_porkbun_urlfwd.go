package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// PORKBUN_URLFWD

func init() {
	Register(TypePORKBUNURLFWD, "PORKBUN_URLFWD", func() dnsv2.RR { return new(PORKBUNURLFWD) }, privatetypesrdata.MakePORKBUNURLFWD)
}

const TypePORKBUNURLFWD = 65321

type PORKBUNURLFWD struct {
	Hdr dnsv2.Header
}

// Typer interface.

func (rr *PORKBUNURLFWD) Type() uint16 { return TypePORKBUNURLFWD }

// RR interface.

func (rr *PORKBUNURLFWD) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *PORKBUNURLFWD) Len() int              { return rr.Hdr.Len() }
func (rr *PORKBUNURLFWD) Data() dnsv2.RDATA {
	return &privatetypesrdata.PORKBUNURLFWD{}
}
func (rr *PORKBUNURLFWD) Clone() dnsv2.RR {
	return &PORKBUNURLFWD{rr.Hdr}
}
func (rr *PORKBUNURLFWD) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tPORKBUN_URLFWD\t" + rr.Data().String()
}

// Parse makes an RDATA for this type using the tokens from dnsv2's parser.
func (rr *PORKBUNURLFWD) Parse(tokens []string, _ string) error {
	return nil
}
