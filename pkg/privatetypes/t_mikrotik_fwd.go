package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// MIKROTIK_FWD

func init() {
	Register(TypeMIKROTIKFWD, "MIKROTIK_FWD", func() dnsv2.RR { return new(MIKROTIKFWD) }, privatetypesrdata.MakeMIKROTIKFWD)
}

const TypeMIKROTIKFWD = 65307

type MIKROTIKFWD struct {
	Hdr dnsv2.Header
}

// Typer interface.

func (rr *MIKROTIKFWD) Type() uint16 { return TypeMIKROTIKFWD }

// RR interface.

func (rr *MIKROTIKFWD) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *MIKROTIKFWD) Len() int              { return rr.Hdr.Len() }
func (rr *MIKROTIKFWD) Data() dnsv2.RDATA {
	return &privatetypesrdata.MIKROTIKFWD{}
}
func (rr *MIKROTIKFWD) Clone() dnsv2.RR {
	return &MIKROTIKFWD{rr.Hdr}
}
func (rr *MIKROTIKFWD) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tMIKROTIK_FWD\t" + rr.Data().String()
}

// Parse makes an RDATA for this type using the tokens from dnsv2's parser.
func (rr *MIKROTIKFWD) Parse(tokens []string, _ string) error {
	return nil
}
