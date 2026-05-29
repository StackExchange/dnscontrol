package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

const TypeADGUARDHOME_AAAA_PASSTHROUGH = 65302

func init() {
	dnsv2.TypeToRR[TypeADGUARDHOME_AAAA_PASSTHROUGH] = func() dnsv2.RR { return new(ADGUARDHOME_AAAA_PASSTHROUGH) }
	dnsv2.TypeToString[TypeADGUARDHOME_AAAA_PASSTHROUGH] = "ADGUARDHOME_AAAA_PASSTHROUGH"
	dnsv2.StringToType["ADGUARDHOME_AAAA_PASSTHROUGH"] = TypeADGUARDHOME_AAAA_PASSTHROUGH
}

// ADGUARDHOME_AAAA_PASSTHROUGH

type ADGUARDHOME_AAAA_PASSTHROUGH struct {
	Hdr dnsv2.Header
}

// Typer interface.
func (rr *ADGUARDHOME_AAAA_PASSTHROUGH) Type() uint16 { return TypeADGUARDHOME_AAAA_PASSTHROUGH }

// RR interface.
func (rr *ADGUARDHOME_AAAA_PASSTHROUGH) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *ADGUARDHOME_AAAA_PASSTHROUGH) Len() int              { return rr.Hdr.Len() }
func (rr *ADGUARDHOME_AAAA_PASSTHROUGH) Data() dnsv2.RDATA {
	return &privatetypesrdata.ADGUARDHOME_AAAA_PASSTHROUGH{}
}
func (rr *ADGUARDHOME_AAAA_PASSTHROUGH) Clone() dnsv2.RR {
	return &ADGUARDHOME_AAAA_PASSTHROUGH{rr.Hdr}
}
func (rr *ADGUARDHOME_AAAA_PASSTHROUGH) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tADGUARDHOME_AAAA_PASSTHROUGH"
}

// Parser interface.
func (rr *ADGUARDHOME_AAAA_PASSTHROUGH) Parse(tokens []string, _ string) error {
	return nil
}
