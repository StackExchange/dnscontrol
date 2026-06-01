package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// AKAMAICDN

func init() {
	Register(TypeAKAMAICDN, "AKAMAICDN",
		func() dnsv2.RR { return new(AKAMAICDN) },
		privatetypesrdata.MakeAKAMAICDN)
}

const TypeAKAMAICDN = 65318

type AKAMAICDN struct {
	Hdr dnsv2.Header
}

// Typer interface.

func (rr *AKAMAICDN) Type() uint16 { return TypeAKAMAICDN }

// RR interface.

func (rr *AKAMAICDN) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *AKAMAICDN) Len() int              { return rr.Hdr.Len() }
func (rr *AKAMAICDN) Data() dnsv2.RDATA {
	return &privatetypesrdata.AKAMAICDN{}
}
func (rr *AKAMAICDN) Clone() dnsv2.RR {
	return &AKAMAICDN{rr.Hdr}
}
func (rr *AKAMAICDN) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tAKAMAICDN\t" + rr.Data().String()
}

// Parse makes an RDATA for this type using the tokens from dnsv2's parser.
func (rr *AKAMAICDN) Parse(tokens []string, _ string) error {
	return nil
}
