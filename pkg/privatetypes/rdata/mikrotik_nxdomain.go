package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
)

type MIKROTIK_NXDOMAIN struct {
	// NXDOMAIN has no data fields — only the .Name matters
}

func (rd MIKROTIK_NXDOMAIN) Len() int {
	return 0
}

func (rd MIKROTIK_NXDOMAIN) String() string {
	return ""
}

func MakeMIKROTIK_NXDOMAIN(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 0 {
		return MIKROTIK_NXDOMAIN{}, fmt.Errorf("MIKROTIK_NXDOMAIN takes no arguments, got %d: %+v", len(args), args)
	}
	return MIKROTIK_NXDOMAIN{}, nil
}
