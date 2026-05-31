package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
)

type BUNNY_DNS_PZ struct {
}

func (rd BUNNY_DNS_PZ) Len() int {
	return 0
}

func (rd BUNNY_DNS_PZ) String() string {
	return ""
}

func MakeBUNNY_DNS_PZ(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 0 {
		return BUNNY_DNS_PZ{}, fmt.Errorf("BUNNY_DNS_PZ requires no arguments, got %d: %+v", len(args), args)
	}
	return BUNNY_DNS_PZ{}, nil
}
