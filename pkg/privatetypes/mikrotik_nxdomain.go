package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// MIKROTIK_NXDOMAIN

func init() {
	Register(TypeMIKROTIK_NXDOMAIN, "MIKROTIK_NXDOMAIN", func() dnsv2.RR { return new(MIKROTIK_NXDOMAIN) }, privatetypesrdata.MakeMIKROTIK_NXDOMAIN)
}

const TypeMIKROTIK_NXDOMAIN = 65308

type MIKROTIK_NXDOMAIN struct {
	Hdr dnsv2.Header
}

// Typer interface.
func (rr *MIKROTIK_NXDOMAIN) Type() uint16 { return TypeMIKROTIK_NXDOMAIN }

// RR interface.
func (rr *MIKROTIK_NXDOMAIN) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *MIKROTIK_NXDOMAIN) Len() int              { return rr.Hdr.Len() }
func (rr *MIKROTIK_NXDOMAIN) Data() dnsv2.RDATA {
	return &privatetypesrdata.MIKROTIK_NXDOMAIN{}
}
func (rr *MIKROTIK_NXDOMAIN) Clone() dnsv2.RR {
	return &MIKROTIK_NXDOMAIN{rr.Hdr}
}
func (rr *MIKROTIK_NXDOMAIN) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tMIKROTIK_NXDOMAIN" // RDATA is empty.
}

// Parser interface.
func (rr *MIKROTIK_NXDOMAIN) Parse(tokens []string, _ string) error {
	return nil
}
