package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v4/pkg/mustbe"
	"github.com/DNSControl/dnscontrol/v4/pkg/txtutil"
)

type MIKROTIKFWD struct {
	ForwardTo string `json:"forward_to"`
}

func (rd MIKROTIKFWD) Len() int {
	return len(rd.ForwardTo) + 1
}

func (rd MIKROTIKFWD) String() string {
	return txtutil.ZoneifyString(rd.ForwardTo)
}

func MakeMIKROTIKFWD(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 1 {
		return MIKROTIKFWD{}, fmt.Errorf("MIKROTKIK_FWD requires 1 argument. Got %d: %+v", len(args), args)
	}
	return MIKROTIKFWD{mustbe.RawString(args[0])}, nil
}
