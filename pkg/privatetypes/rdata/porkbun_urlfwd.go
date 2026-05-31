package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
)

type PORKBUN_URLFWD struct {
	// Deprecated.  Leaving this empty for now. I think the provider substiutes the replacement (URL or URL301) in the provider code, so we don't need to do anything here.
}

func (rd PORKBUN_URLFWD) Len() int {
	return 0
}

func (rd PORKBUN_URLFWD) String() string {
	return "PORKBUN_URLFWD.String() should not be called"
}

func MakePORKBUN_URLFWD(origin string, args ...any) (dnsv2.RDATA, error) {
	return PORKBUN_URLFWD{}, fmt.Errorf("MakePORKBUN_URLFWD() should not be used")
}
