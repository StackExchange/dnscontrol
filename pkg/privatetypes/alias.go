package privatetypes

import (
	"strconv"

	dnsv2 "codeberg.org/miekg/dns"
	dnsutilv2 "codeberg.org/miekg/dns/dnsutil"
	privatetypesrdata "github.com/DNSControl/dnscontrol/v4/pkg/privatetypes/rdata"
)

func init() {
	dnsv2.TypeToRR[TypeALIAS] = func() dnsv2.RR { return new(ALIAS) }
	dnsv2.TypeToString[TypeALIAS] = "ALIAS"
	dnsv2.StringToType["ALIAS"] = TypeALIAS
}

// ALIAS

type ALIAS struct {
	Hdr    dnsv2.Header
	Target string
}

const TypeALIAS = 65300

// Typer interface.
func (rr *ALIAS) Type() uint16 { return TypeALIAS }

// RR interface.
func (rr *ALIAS) Header() *dnsv2.Header { return &rr.Hdr }
func (rr *ALIAS) Len() int              { return rr.Hdr.Len() + 1 + len(rr.Target) }
func (rr *ALIAS) Data() dnsv2.RDATA     { return &privatetypesrdata.ALIAS{Target: rr.Target} }
func (rr *ALIAS) Clone() dnsv2.RR       { return &ALIAS{rr.Hdr, rr.Target} }
func (rr *ALIAS) String() string {
	return rr.Header().Name + "\t" +
		strconv.FormatInt(int64(rr.Header().TTL), 10) + "\t" +
		dnsutilv2.ClassToString(rr.Header().Class) + "\tALIAS\t" +
		rr.Target
}

// Parser interface.
func (rr *ALIAS) Parse(tokens []string, _ string) error {
	if len(tokens) < 1 { // no rdata
		return nil
	}
	rr.Target = tokens[0]
	return nil
}
