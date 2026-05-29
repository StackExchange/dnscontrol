package privatetypes

import (
	"fmt"
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// R53_ALIAS

func init() {
	Register(TypeR53_ALIAS, "R53_ALIAS", func() dnsv2.RR { return new(R53_ALIAS) })
}

const TypeR53_ALIAS = 65306

type R53_ALIAS struct {
	Hdr dnsv2.Header

	AliasType, Target, EvalTargetHealth string
}

// Typer interface.
func (rr *R53_ALIAS) Type() uint16 { return TypeR53_ALIAS }

// RR interface.
func (rr *R53_ALIAS) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *R53_ALIAS) Len() int {
	return rr.Hdr.Len() +
		1 + len(rr.AliasType) +
		1 + len(rr.Target) +
		1 + len(rr.EvalTargetHealth)
}
func (rr *R53_ALIAS) Data() dnsv2.RDATA { return &privatetypesrdata.R53_ALIAS{Target: rr.Target} }
func (rr *R53_ALIAS) Clone() dnsv2.RR {
	return &R53_ALIAS{rr.Hdr, rr.AliasType, rr.Target, rr.EvalTargetHealth}
}
func (rr *R53_ALIAS) String() string {
	return (rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tR53_ALIAS\t" +
		" " + rr.AliasType +
		" " + rr.Target +
		" " + rr.EvalTargetHealth)
}

// Parser interface.
func (rr *R53_ALIAS) Parse(tokens []string, s string) error {
	args := TokensToArgs(tokens)
	if len(args) != 3 {
		return fmt.Errorf("%s requires exactly 3 arguments, got %d", dnsutilv2.TypeToString(rr.Type()), len(args))
	}
	rr.AliasType = args[0]
	rr.Target = args[1]
	rr.EvalTargetHealth = args[2]
	return nil
}
