package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// MIKROTIK_NXDOMAIN

func init() {
	Register(TypeMIKROTIKNXDOMAIN, "MIKROTIK_NXDOMAIN", func() dnsv2.RR { return new(MIKROTIKNXDOMAIN) }, privatetypesrdata.MakeMIKROTIKNXDOMAIN)
}

const TypeMIKROTIKNXDOMAIN = 65308

type MIKROTIKNXDOMAIN struct {
	Hdr dnsv2.Header
}

// Typer interface.

func (rr *MIKROTIKNXDOMAIN) Type() uint16 { return TypeMIKROTIKNXDOMAIN }

// RR interface.

func (rr *MIKROTIKNXDOMAIN) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *MIKROTIKNXDOMAIN) Len() int              { return rr.Hdr.Len() }
func (rr *MIKROTIKNXDOMAIN) Data() dnsv2.RDATA {
	return &privatetypesrdata.MIKROTIKNXDOMAIN{}
}
func (rr *MIKROTIKNXDOMAIN) Clone() dnsv2.RR {
	return &MIKROTIKNXDOMAIN{rr.Hdr}
}
func (rr *MIKROTIKNXDOMAIN) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tMIKROTIK_NXDOMAIN" // RDATA is empty.
}

// Parse makes an RDATA for this type using the tokens from dnsv2's parser.
func (rr *MIKROTIKNXDOMAIN) Parse(tokens []string, _ string) error {
	return nil
}
