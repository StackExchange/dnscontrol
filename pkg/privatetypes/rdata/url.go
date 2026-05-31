package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v4/pkg/mustbe"
	"github.com/DNSControl/dnscontrol/v4/pkg/txtutil"
)

type URL struct {
	Location string

	// Pornbun-specific fields:
	PorkbunIncludePath bool
	PorkbunWildCard    bool
}

func (rd URL) Len() int {
	return len(rd.Location) + 3
}

func (rd URL) String() string {
	return fmt.Sprintf("%s %t %t", txtutil.ZoneifyString(rd.Location), rd.PorkbunIncludePath, rd.PorkbunWildCard)
}

// MakeURL creates an RDATA from args.
func MakeURL(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 3 {
		return URL{}, fmt.Errorf("URL expects 3 arguments, got %d: %+v", len(args), args)
	}
	return URL{mustbe.RawString(args[0]), mustbe.Bool(args[1]), mustbe.Bool(args[2])}, nil
}
