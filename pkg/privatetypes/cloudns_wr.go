package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// CLOUDNS_WR

func init() {
	Register(TypeCLOUDNS_WR, "CLOUDNS_WR", func() dnsv2.RR { return new(CLOUDNS_WR) }, privatetypesrdata.MakeCLOUDNS_WR)
}

const TypeCLOUDNS_WR = 65315

type CLOUDNS_WR struct {
	Hdr dnsv2.Header
}

// Typer interface.
func (rr *CLOUDNS_WR) Type() uint16 { return TypeCLOUDNS_WR }

// RR interface.
func (rr *CLOUDNS_WR) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *CLOUDNS_WR) Len() int              { return rr.Hdr.Len() }
func (rr *CLOUDNS_WR) Data() dnsv2.RDATA {
	return &privatetypesrdata.CLOUDNS_WR{}
}
func (rr *CLOUDNS_WR) Clone() dnsv2.RR {
	return &CLOUDNS_WR{rr.Hdr}
}
func (rr *CLOUDNS_WR) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tCLOUDNS_WR\t" + rr.Data().String()
}

// Parser interface.
func (rr *CLOUDNS_WR) Parse(tokens []string, _ string) error {
	return nil
}
