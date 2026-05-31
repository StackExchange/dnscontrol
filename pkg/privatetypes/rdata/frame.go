package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v4/pkg/mustbe"
	"github.com/DNSControl/dnscontrol/v4/pkg/txtutil"
)

type FRAME struct {
	Target string
}

func (rd FRAME) Len() int {
	return len(rd.Target) + 1
}

func (rd FRAME) String() string {
	return txtutil.ZoneifyString(rd.Target)
}

func MakeFRAME(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 1 {
		return FRAME{}, fmt.Errorf("FRAME requires exactly 1 argument, got %d: %+v", len(args), args)
	}
	return FRAME{Target: mustbe.RawString(args[0])}, nil
}
