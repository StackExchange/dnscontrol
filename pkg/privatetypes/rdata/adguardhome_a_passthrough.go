package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
)

type ADGUARDHOME_A_PASSTHROUGH struct {
}

func (rd ADGUARDHOME_A_PASSTHROUGH) Len() int {
	return 0
}

func (rd ADGUARDHOME_A_PASSTHROUGH) String() string {
	return ""
}

func MakeADGUARDHOME_A_PASSTHROUGH(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 0 {
		return ADGUARDHOME_A_PASSTHROUGH{}, fmt.Errorf("ADGUARDHOME_A_PASSTHROUGH expects 0 arguments, got %d: %+v", len(args), args)
	}
	return ADGUARDHOME_A_PASSTHROUGH{}, nil
}
