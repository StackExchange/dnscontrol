package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
)

type AKAMAICDN struct {
}

func (rd AKAMAICDN) Len() int {
	return 0
}

func (rd AKAMAICDN) String() string {
	return ""
}

func MakeAKAMAICDN(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 0 {
		return AKAMAICDN{}, fmt.Errorf("AKAMAICDN requires exactly 0 arguments")
	}
	return AKAMAICDN{}, nil
}
