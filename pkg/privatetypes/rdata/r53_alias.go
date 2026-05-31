package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v4/pkg/mustbe"
	"github.com/DNSControl/dnscontrol/v4/pkg/txtutil"
)

type R53_ALIAS struct {
	AliasType        string
	Target           string
	EvalTargetHealth string
	ZoneID           string
}

func (rr R53_ALIAS) Len() int {
	return len(rr.AliasType) +
		1 + len(rr.Target) +
		1 + len(rr.EvalTargetHealth) +
		1 + len(rr.ZoneID)
}

func (rd R53_ALIAS) String() string {
	return txtutil.Zoneify([]string{rd.AliasType, rd.Target, rd.EvalTargetHealth, rd.ZoneID})
}

func MakeR53_ALIAS(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 4 {
		return R53_ALIAS{}, fmt.Errorf("R53_ALIAS expects 4 arguments, got %d: %+v", len(args), args)
	}
	return R53_ALIAS{mustbe.RawString(args[0]), mustbe.RawString(args[1]), mustbe.RawString(args[2]), mustbe.RawString(args[3])}, nil
	// TODO(tlim): Could these be validated more? For example, the first argument should be one of "A", "AAAA", "CNAME", "MX", "NS", "PTR", "SPF", "SRV", or "TXT". The third argument should be either "true" or "false".
}
