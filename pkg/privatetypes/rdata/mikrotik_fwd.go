package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v4/pkg/mustbe"
	"github.com/DNSControl/dnscontrol/v4/pkg/txtutil"
)

type MIKROTIK_FWD struct {
	ForwardTo string `json:"forward_to"`
}

func (rd MIKROTIK_FWD) Len() int {
	return len(rd.ForwardTo) + 1
}

func (rd MIKROTIK_FWD) String() string {
	return txtutil.ZoneifyString(rd.ForwardTo)
}

func MakeMIKROTIK_FWD(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 1 {
		return MIKROTIK_FWD{}, fmt.Errorf("MIKROTKIK_FWD requires 1 argument. Got %d: %+v", len(args), args)
	}
	return MIKROTIK_FWD{mustbe.RawString(args[0])}, nil
}
