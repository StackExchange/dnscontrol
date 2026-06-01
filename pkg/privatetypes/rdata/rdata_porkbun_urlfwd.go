package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
)

type PORKBUNURLFWD struct {
	// Deprecated.  Leaving this empty for now. I think the provider substiutes the replacement (URL or URL301) in the provider code, so we don't need to do anything here.
}

func (rd PORKBUNURLFWD) Len() int {
	return 0
}

func (rd PORKBUNURLFWD) String() string {
	return "PORKBUNURLFWD.String() should not be called"
}

func MakePORKBUNURLFWD(origin string, args ...any) (dnsv2.RDATA, error) {
	return PORKBUNURLFWD{}, fmt.Errorf("MakePORKBUN_URLFWD() should not be used")
}
