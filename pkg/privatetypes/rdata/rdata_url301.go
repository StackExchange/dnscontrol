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
	PorkbunIncludePath bool `json:"porkbun_include_path"`
	PorkbunWildCard    bool `json:"porkbun_wildcard"`
}

func (rd URL301) Len() int {
	return len(rd.Location) + 1
}

func (rd URL301) String() string {
	return txtutil.ZoneifyString(rd.Location) + fmt.Sprintf(" %t %t", rd.PorkbunIncludePath, rd.PorkbunWildCard)
}

func MakeURL301(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 3 {
		return URL301{}, fmt.Errorf("URL301 expects 3 arguments, got %d: %+v", len(args), args)
	}
	return URL301{Location: mustbe.TargetHost(origin, args[0]), PorkbunIncludePath: mustbe.Bool(args[1]), PorkbunWildCard: mustbe.Bool(args[2])}, nil
}
