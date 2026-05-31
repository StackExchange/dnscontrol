package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
)

type ADGUARDHOME_AAAA_PASSTHROUGH struct {
}

func (rd ADGUARDHOME_AAAA_PASSTHROUGH) Len() int {
	return 0
}

func (rd ADGUARDHOME_AAAA_PASSTHROUGH) String() string {
	return ""
}

func MakeADGUARDHOME_AAAA_PASSTHROUGH(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 0 {
		return ADGUARDHOME_AAAA_PASSTHROUGH{}, fmt.Errorf("ADGUARDHOME_AAAA_PASSTHROUGH requires 0 arguments. Got %d: %+v", len(args), args)
	}
	return ADGUARDHOME_AAAA_PASSTHROUGH{}, nil
}
