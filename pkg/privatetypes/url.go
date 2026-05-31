package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// URL

func init() {
	Register(TypeURL, "URL", func() dnsv2.RR { return new(URL) }, privatetypesrdata.MakeURL)
}

const TypeURL = 65310

type URL struct {
	Hdr dnsv2.Header
}

// Typer interface.

func (rr *URL) Type() uint16 { return TypeURL }

// RR interface.

func (rr *URL) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *URL) Len() int              { return rr.Hdr.Len() }
func (rr *URL) Data() dnsv2.RDATA {
	return &privatetypesrdata.URL{}
}
func (rr *URL) Clone() dnsv2.RR {
	return &URL{rr.Hdr}
}
func (rr *URL) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tURL\t" + rr.Data().String()
}

// Parse makes an RDATA for this type using the tokens from dnsv2's parser.
func (rr *URL) Parse(tokens []string, _ string) error {
	return nil
}
