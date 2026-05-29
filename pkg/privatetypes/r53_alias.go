package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

func init() {
	dnsv2.TypeToRR[TypeR53_ALIAS] = func() dnsv2.RR { return new(R53_ALIAS) }
	dnsv2.TypeToString[TypeR53_ALIAS] = "R53_ALIAS"
	dnsv2.StringToType["R53_ALIAS"] = TypeR53_ALIAS
}

// R53_ALIAS

type R53_ALIAS struct {
	Hdr dnsv2.Header

	AliasType, Target, EvalTargetHealth string
}

const TypeR53_ALIAS = 65302

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
	if len(tokens) < 3 { // no rdata
		return nil
	}
	rr.AliasType = tokens[0]
	rr.Target = tokens[1]
	rr.EvalTargetHealth = tokens[2]
	return nil
}
