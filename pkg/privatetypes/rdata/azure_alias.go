package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v4/pkg/mustbe"
)

type AZUREALIAS struct {
	AliasType string
	Target    string
}

func (rd AZUREALIAS) Len() int {
	return len(rd.Target) + 1 + len(rd.AliasType)
}

func (rd AZUREALIAS) String() string {
	return rd.AliasType + " " + rd.Target
}

func MakeAZUREALIAS(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 2 {
		return AZUREALIAS{}, fmt.Errorf("AZURE_ALIAS requires no arguments, got %d: %+v", len(args), args)
	}
	return AZUREALIAS{mustbe.RawString(args[0]), mustbe.Host(origin, args[1])}, nil
}
