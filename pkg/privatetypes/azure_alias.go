package privatetypes

import (
	"fmt"
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

// AZURE_ALIAS

func init() {
	Register(TypeAZUREALIAS, "AZURE_ALIAS", func() dnsv2.RR { return new(AZUREALIAS) }, privatetypesrdata.MakeAZUREALIAS)
}

const TypeAZUREALIAS = 65304

type AZUREALIAS struct {
	Hdr       dnsv2.Header
	AliasType string
	Target    string
}

// Typer interface.

func (rr *AZUREALIAS) Type() uint16 { return TypeAZUREALIAS }

// RR interface.

func (rr *AZUREALIAS) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *AZUREALIAS) Len() int              { return rr.Hdr.Len() + rr.Data().Len() }
func (rr *AZUREALIAS) Data() dnsv2.RDATA {
	return &privatetypesrdata.AZUREALIAS{AliasType: rr.AliasType, Target: rr.Target}
}
func (rr *AZUREALIAS) Clone() dnsv2.RR {
	return &AZUREALIAS{Hdr: rr.Hdr, AliasType: rr.AliasType, Target: rr.Target}
}
func (rr *AZUREALIAS) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tAZURE_ALIAS\t" + rr.Data().String()
}

// Parse makes an RDATA for this type using the tokens from dnsv2's parser.
func (rr *AZUREALIAS) Parse(tokens []string, _ string) error {
	args := TokensToArgs(tokens)
	if len(args) != 1 {
		return fmt.Errorf("%s requires exactly 1 argument, got %d", dnsutilv2.TypeToString(rr.Type()), len(args))
	}
	rr.Target = args[0]
	return nil
}
