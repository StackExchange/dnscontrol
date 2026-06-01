package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
)

type BUNNYDNSPZ struct {
}

func (rd BUNNYDNSPZ) Len() int {
	return 0
}

func (rd BUNNYDNSPZ) String() string {
	return ""
}

func MakeBUNNYDNSPZ(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 0 {
		return BUNNYDNSPZ{}, fmt.Errorf("BUNNY_DNS_PZ requires no arguments, got %d: %+v", len(args), args)
	}
	return BUNNYDNSPZ{}, nil
}
