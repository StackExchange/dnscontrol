package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
)

type BUNNY_DNS_RDR struct {
}

func (rd BUNNY_DNS_RDR) Len() int {
	return 0
}

func (rd BUNNY_DNS_RDR) String() string {
	return ""
}

func MakeBUNNY_DNS_RDR(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 0 {
		return BUNNY_DNS_RDR{}, fmt.Errorf("BUNNY_DNS_RDR requires no arguments, got %d: %+v", len(args), args)
	}
	return BUNNY_DNS_RDR{}, nil
}
