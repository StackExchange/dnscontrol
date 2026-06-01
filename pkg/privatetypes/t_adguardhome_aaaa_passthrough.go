package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// ADGUARDHOME_AAAA_PASSTHROUGH

func init() {
	Register(TypeADGUARDHOMEAAAAPASSTHROUGH, "ADGUARDHOME_AAAA_PASSTHROUGH", func() dnsv2.RR { return new(ADGUARDHOMEAAAAPASSTHROUGH) }, privatetypesrdata.MakeADGUARDHOMEAAAAPASSTHROUGH)
}

const TypeADGUARDHOMEAAAAPASSTHROUGH = 65302

type ADGUARDHOMEAAAAPASSTHROUGH struct {
	Hdr dnsv2.Header
}

// Typer interface.

func (rr *ADGUARDHOMEAAAAPASSTHROUGH) Type() uint16 { return TypeADGUARDHOMEAAAAPASSTHROUGH }

// RR interface.

func (rr *ADGUARDHOMEAAAAPASSTHROUGH) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *ADGUARDHOMEAAAAPASSTHROUGH) Len() int              { return rr.Hdr.Len() }
func (rr *ADGUARDHOMEAAAAPASSTHROUGH) Data() dnsv2.RDATA {
	return &privatetypesrdata.ADGUARDHOMEAAAAPASSTHROUGH{}
}
func (rr *ADGUARDHOMEAAAAPASSTHROUGH) Clone() dnsv2.RR {
	return &ADGUARDHOMEAAAAPASSTHROUGH{rr.Hdr}
}
func (rr *ADGUARDHOMEAAAAPASSTHROUGH) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tADGUARDHOME_AAAA_PASSTHROUGH\t" + rr.Data().String()
}

// Parse makes an RDATA for this type using the tokens from dnsv2's parser.
func (rr *ADGUARDHOMEAAAAPASSTHROUGH) Parse(tokens []string, _ string) error {
	return nil
}
