package privatetypes

import (
	"fmt"
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
	"github.com/DNSControl/dnscontrol/v4/pkg/txtutil"
)

// CFWORKERROUTE

func init() {
	Register(TypeCFWORKERROUTE, "CF_WORKER_ROUTE", func() dnsv2.RR { return new(CFWORKERROUTE) })
}

const TypeCFWORKERROUTE = 65305

type CFWORKERROUTE struct {
	Hdr  dnsv2.Header
	When string
	Then string
}

// Typer interface.
func (rr *CFWORKERROUTE) Type() uint16 { return TypeCFWORKERROUTE }

// RR interface.
func (rr *CFWORKERROUTE) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *CFWORKERROUTE) Len() int              { return rr.Hdr.Len() + 1 + len(rr.When) + 1 + len(rr.Then) }
func (rr *CFWORKERROUTE) Data() dnsv2.RDATA {
	return &privatetypesrdata.CFWORKERROUTE{When: rr.When, Then: rr.Then}
}
func (rr *CFWORKERROUTE) Clone() dnsv2.RR { return &CFWORKERROUTE{rr.Hdr, rr.When, rr.Then} }
func (rr *CFWORKERROUTE) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tCF_WORKER_ROUTE\t" +
		txtutil.Zoneify([]string{rr.When, rr.Then})
}

// Parser interface.
func (rr *CFWORKERROUTE) Parse(tokens []string, s string) error {
	args := TokensToArgs(tokens)
	if len(args) != 2 {
		return fmt.Errorf("%s requires exactly 2 arguments, got %d", dnsutilv2.TypeToString(rr.Type()), len(args))
	}
	rr.When = args[0]
	rr.Then = args[1]
	return nil
}
