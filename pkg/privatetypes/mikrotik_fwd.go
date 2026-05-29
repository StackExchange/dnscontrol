package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// MIKROTIK_FWD

func init() {
	Register(TypeMIKROTIK_FWD, "MIKROTIK_FWD", func() dnsv2.RR { return new(MIKROTIK_FWD) })
}

const TypeMIKROTIK_FWD = 65307

type MIKROTIK_FWD struct {
	Hdr dnsv2.Header
}

// Typer interface.
func (rr *MIKROTIK_FWD) Type() uint16 { return TypeMIKROTIK_FWD }

// RR interface.
func (rr *MIKROTIK_FWD) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *MIKROTIK_FWD) Len() int              { return rr.Hdr.Len() }
func (rr *MIKROTIK_FWD) Data() dnsv2.RDATA {
	return &privatetypesrdata.MIKROTIK_FWD{}
}
func (rr *MIKROTIK_FWD) Clone() dnsv2.RR {
	return &MIKROTIK_FWD{rr.Hdr}
}
func (rr *MIKROTIK_FWD) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tMIKROTIK_FWD"
}

// Parser interface.
func (rr *MIKROTIK_FWD) Parse(tokens []string, _ string) error {
	return nil
}
