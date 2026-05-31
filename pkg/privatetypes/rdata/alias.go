package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v4/pkg/mustbe"
)

type ALIAS struct {
	Target string
}

func (rd ALIAS) Len() int {
	return len(rd.Target) + 1
}

func (rd ALIAS) String() string {
	return rd.Target
}

func MakeALIAS(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 1 {
		return ALIAS{}, fmt.Errorf("ALIAS requires 1 argument, got %d: %+v", len(args), args)
	}
	return ALIAS{mustbe.Host(origin, args[0])}, nil
}
