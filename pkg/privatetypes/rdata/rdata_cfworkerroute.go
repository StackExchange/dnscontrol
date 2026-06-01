package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v4/pkg/mustbe"
	"github.com/DNSControl/dnscontrol/v4/pkg/txtutil"
)

type CFWORKERROUTE struct {
	When string
	Then string
}

func (rd CFWORKERROUTE) Len() int {
	return len(rd.When) + 1 + len(rd.Then)
}

func (rd CFWORKERROUTE) String() string {
	return txtutil.ZoneifyQuoted([]string{rd.When, rd.Then})
}

func MakeCFWORKERROUTE(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 2 {
		return CFWORKERROUTE{}, fmt.Errorf("CFWORKERROUTE requires exactly 2 arguments, got %d: %+v", len(args), args)
	}
	return CFWORKERROUTE{mustbe.RawString(args[0]), mustbe.RawString(args[1])}, nil
}
