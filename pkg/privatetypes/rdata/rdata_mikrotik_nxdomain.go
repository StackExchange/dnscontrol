package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
)

type MIKROTIKNXDOMAIN struct {
	// NXDOMAIN has no data fields — only the .Name matters
}

func (rd MIKROTIKNXDOMAIN) Len() int {
	return 0
}

func (rd MIKROTIKNXDOMAIN) String() string {
	return ""
}

// MakeMIKROTIKNXDOMAIN creates an RDATA from args.
func MakeMIKROTIKNXDOMAIN(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 0 {
		return MIKROTIKNXDOMAIN{}, fmt.Errorf("MIKROTIK_NXDOMAIN takes no arguments, got %d: %+v", len(args), args)
	}
	return MIKROTIKNXDOMAIN{}, nil
}
