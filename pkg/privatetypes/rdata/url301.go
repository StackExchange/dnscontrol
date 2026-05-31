package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v4/pkg/mustbe"
	"github.com/DNSControl/dnscontrol/v4/pkg/txtutil"
)

type URL301 struct {
	Location string

	// Pornbun-specific fields:
	Porkbun_IncludePath bool `json:"porkbun_include_path"`
	Porkbun_WildCard    bool `json:"porkbun_wildcard"`
}

func (rd URL301) Len() int {
	return len(rd.Location) + 1
}

func (rd URL301) String() string {
	return txtutil.ZoneifyString(rd.Location) + fmt.Sprintf(" %t %t", rd.Porkbun_IncludePath, rd.Porkbun_WildCard)
}

func MakeURL301(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 3 {
		return URL301{}, fmt.Errorf("URL301 expects 3 arguments, got %d: %+v", len(args), args)
	}
	return URL301{Location: mustbe.Host(origin, args[0]), Porkbun_IncludePath: mustbe.Bool(args[1]), Porkbun_WildCard: mustbe.Bool(args[2])}, nil
}
