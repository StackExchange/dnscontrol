package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

const TypeAZURE_ALIAS = 65304

func init() {
	dnsv2.TypeToRR[TypeAZURE_ALIAS] = func() dnsv2.RR { return new(AZURE_ALIAS) }
	dnsv2.TypeToString[TypeAZURE_ALIAS] = "AZURE_ALIAS"
	dnsv2.StringToType["AZURE_ALIAS"] = TypeAZURE_ALIAS
}

// AZURE_ALIAS

type AZURE_ALIAS struct {
	Hdr       dnsv2.Header
	AliasType string
	Target    string
}

// Typer interface.
func (rr *AZURE_ALIAS) Type() uint16 { return TypeAZURE_ALIAS }

// RR interface.
func (rr *AZURE_ALIAS) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *AZURE_ALIAS) Len() int              { return rr.Hdr.Len() + rr.Data().Len() }
func (rr *AZURE_ALIAS) Data() dnsv2.RDATA {
	return &privatetypesrdata.AZURE_ALIAS{AliasType: rr.AliasType, Target: rr.Target}
}
func (rr *AZURE_ALIAS) Clone() dnsv2.RR {
	return &AZURE_ALIAS{Hdr: rr.Hdr, AliasType: rr.AliasType, Target: rr.Target}
}
func (rr *AZURE_ALIAS) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tAZURE_ALIAS\t" +
		rr.Data().String()
}

// Parser interface.
func (rr *AZURE_ALIAS) Parse(tokens []string, _ string) error {
	if len(tokens) < 1 { // no rdata
		return nil
	}
	rr.Target = tokens[0]
	return nil
}
