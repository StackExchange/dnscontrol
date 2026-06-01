package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
)

type AKAMAITLC struct {
}

func (rd AKAMAITLC) Len() int {
	return 0
}

func (rd AKAMAITLC) String() string {
	return ""
}

func MakeAKAMAITLC(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 0 {
		return AKAMAITLC{}, fmt.Errorf("AKAMAITLC takes no arguments")
	}
	return AKAMAITLC{}, nil
}
