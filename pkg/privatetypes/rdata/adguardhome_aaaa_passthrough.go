package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
)

type ADGUARDHOMEAAAAPASSTHROUGH struct {
}

func (rd ADGUARDHOMEAAAAPASSTHROUGH) Len() int {
	return 0
}

func (rd ADGUARDHOMEAAAAPASSTHROUGH) String() string {
	return ""
}

func MakeADGUARDHOMEAAAAPASSTHROUGH(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 0 {
		return ADGUARDHOMEAAAAPASSTHROUGH{}, fmt.Errorf("ADGUARDHOME_AAAA_PASSTHROUGH requires 0 arguments. Got %d: %+v", len(args), args)
	}
	return ADGUARDHOMEAAAAPASSTHROUGH{}, nil
}
