package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v4/pkg/mustbe"
	"github.com/DNSControl/dnscontrol/v4/pkg/txtutil"
)

type CLOUDNS_WR struct {
	Target string
}

func (rd CLOUDNS_WR) Len() int {
	return 0
}

func (rd CLOUDNS_WR) String() string {
	return txtutil.ZoneifyString(rd.Target)
}

func MakeCLOUDNS_WR(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 1 {
		return CLOUDNS_WR{}, fmt.Errorf("CLOUDNS_WR requires exactly 1 argument, got %d: %+v", len(args), args)
	}
	return CLOUDNS_WR{Target: mustbe.RawString(args[0])}, nil
}
