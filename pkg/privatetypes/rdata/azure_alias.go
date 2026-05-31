package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v4/pkg/mustbe"
)

type AZURE_ALIAS struct {
	AliasType string
	Target    string
}

func (rd AZURE_ALIAS) Len() int {
	return len(rd.Target) + 1 + len(rd.AliasType)
}

func (rd AZURE_ALIAS) String() string {
	return rd.AliasType + " " + rd.Target
}

func MakeAZURE_ALIAS(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 2 {
		return AZURE_ALIAS{}, fmt.Errorf("AZURE_ALIAS requires no arguments, got %d: %+v", len(args), args)
	}
	return AZURE_ALIAS{mustbe.RawString(args[0]), mustbe.Host(origin, args[1])}, nil
}
