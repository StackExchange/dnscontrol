package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// URL301

func init() {
	Register(TypeURL301, "URL301", func() dnsv2.RR { return new(URL301) }, privatetypesrdata.MakeURL301)
}

const TypeURL301 = 65311

type URL301 struct {
	Hdr dnsv2.Header
}

// Typer interface.

func (rr *URL301) Type() uint16 { return TypeURL301 }

// RR interface.

func (rr *URL301) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *URL301) Len() int              { return rr.Hdr.Len() }
func (rr *URL301) Data() dnsv2.RDATA {
	return &privatetypesrdata.URL301{}
}
func (rr *URL301) Clone() dnsv2.RR {
	return &URL301{rr.Hdr}
}
func (rr *URL301) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tURL301\t" + rr.Data().String()
}

// Parse makes an RDATA for this type using the tokens from dnsv2's parser.
func (rr *URL301) Parse(tokens []string, _ string) error {
	return nil
}
