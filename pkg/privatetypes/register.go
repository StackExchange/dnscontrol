package privatetypes

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
)

// TypeToMakeRDATA returns a function that accepts arguments of any type and returns a dnsv2.RDATA struct.
// Examples:
//
//	demoRC1, err := TypeToMakeRDATA[dnsv2.TypeA](mustbe.IPv4("1.2.3.4"))
//	demoRC2, err := TypeToMakeRDATA[dnsv2.TypeCNAME](mustbe.Host("www", "example.com"))
//	demoRC3, err := TypeToMakeRDATA[privatetypes.TypeCFWORKERROUTE](mustbe.RawString("example.com/*"), mustbe.RawString("example.com/worker"))
var TypeToMakeRDATA = make(map[uint16]func(origin string, args ...any) (dnsv2.RDATA, error))

// Register registers a new private RR type. It panics if the code point or name is already in use.
func Register(codepoint uint16, name string, newFn func() dnsv2.RR, makeFn func(origin string, args ...any) (dnsv2.RDATA, error)) {

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

	// typenum -> func() RR  i.e. a function that creates a new RR struct for the given code point.
	if dnsv2.TypeToRR[codepoint] != nil {
		panic(fmt.Sprintf("TypeToRR[%d] already in use", codepoint))
	}
	dnsv2.TypeToRR[codepoint] = newFn

	// typenum -> typename
	if dnsv2.TypeToString[codepoint] != "" {
		panic(fmt.Sprintf("TypeToString[%d] already in use by %s", codepoint, dnsv2.TypeToString[codepoint]))
	}
	dnsv2.TypeToString[codepoint] = name

	// typename -> typenum
	if s, exists := dnsv2.StringToType[name]; exists {
		panic(fmt.Sprintf("StringToType[%s] already in use by %d", name, s))
	}
	dnsv2.StringToType[name] = codepoint

	// typenum -> func(args ...any) (RDATA, error) i.e. a function that creates an RDATA struct for the given code point, with fields filled from the given args.
	if s, exists := TypeToMakeRDATA[codepoint]; exists {
		panic(fmt.Sprintf("TypeToMakeRDATA[%s] already in use by %d", name, s))
	}
	TypeToMakeRDATA[codepoint] = makeFn
}
