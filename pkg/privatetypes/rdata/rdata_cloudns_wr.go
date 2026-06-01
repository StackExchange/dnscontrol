package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v4/pkg/mustbe"
	"github.com/DNSControl/dnscontrol/v4/pkg/txtutil"
)

type CLOUDNSWR struct {
	Target string
}

func (rd CLOUDNSWR) Len() int {
	return 0
}

func (rd CLOUDNSWR) String() string {
	return txtutil.ZoneifyString(rd.Target)
}

func MakeCLOUDNSWR(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 1 {
		return CLOUDNSWR{}, fmt.Errorf("CLOUDNS_WR requires exactly 1 argument, got %d: %+v", len(args), args)
	}
	return CLOUDNSWR{Target: mustbe.RawString(args[0])}, nil
}
