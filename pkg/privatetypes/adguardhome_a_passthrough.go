package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// ADGUARDHOME_A_PASSTHROUGH

func init() {
	Register(TypeADGUARDHOME_A_PASSTHROUGH, "ADGUARDHOME_A_PASSTHROUGH", func() dnsv2.RR { return new(ADGUARDHOME_A_PASSTHROUGH) })
}

const TypeADGUARDHOME_A_PASSTHROUGH = 65301

type ADGUARDHOME_A_PASSTHROUGH struct {
	Hdr dnsv2.Header
}

// Typer interface.
func (rr *ADGUARDHOME_A_PASSTHROUGH) Type() uint16 { return TypeADGUARDHOME_A_PASSTHROUGH }

// RR interface.
func (rr *ADGUARDHOME_A_PASSTHROUGH) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *ADGUARDHOME_A_PASSTHROUGH) Len() int              { return rr.Hdr.Len() }
func (rr *ADGUARDHOME_A_PASSTHROUGH) Data() dnsv2.RDATA {
	return &privatetypesrdata.ADGUARDHOME_A_PASSTHROUGH{}
}
func (rr *ADGUARDHOME_A_PASSTHROUGH) Clone() dnsv2.RR { return &ADGUARDHOME_A_PASSTHROUGH{rr.Hdr} }
func (rr *ADGUARDHOME_A_PASSTHROUGH) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tADGUARDHOME_A_PASSTHROUGH"
}

// Parser interface.
func (rr *ADGUARDHOME_A_PASSTHROUGH) Parse(tokens []string, _ string) error {
	return nil
}
