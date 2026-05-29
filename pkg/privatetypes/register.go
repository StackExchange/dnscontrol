package privatetypes

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
)

/*
# Private Resource Records

Any struct can be used as a private resource record. To make it work you need to implement the following interfaces.

  - [Typer], to give your RR a code point, and see documentation of that interface.
  - [RR], all RRs implement this, if you want to have a private EDNS0 option, implement [EDNS0] interface, this
    adds a Pseudo() bool method.
  - [Parser], so it can be parsed to and from strings.
  - [Packer], if you need to use your new RR on the wire.
  - [Comparer], if your RR will be signed with DNSSEC.

See rr_test.go for a complete example for both an external [RR] and [EDNS0].
*/

func Register(codepoint uint16, name string, newFn func() dnsv2.RR) {

	if dnsv2.TypeToRR[codepoint] != nil {
		panic(fmt.Sprintf("TypeToRR[%d] already in use", codepoint))
	}
	dnsv2.TypeToRR[codepoint] = newFn

	if dnsv2.TypeToString[codepoint] != "" {
		panic(fmt.Sprintf("TypeToString[%d] already in use by %s", codepoint, dnsv2.TypeToString[codepoint]))
	}
	dnsv2.TypeToString[codepoint] = name

	if s, exists := dnsv2.StringToType[name]; exists {
		panic(fmt.Sprintf("StringToType[%s] already in use by %d", name, s))
	}
	dnsv2.StringToType[name] = codepoint
}
