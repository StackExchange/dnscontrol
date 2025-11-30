package rtype

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
	"github.com/miekg/dns"
)

func init() {
	rtypecontrol.Register(&RP{})
}

// RP RR. See RFC 1138, Section 2.2.
type RP struct {
	dns.RP
}

func (handle *RP) Name() string {
	return "RP"
}

func (handle *RP) FromArgs(dc *models.DomainConfig, rec *models.RecordConfig, args []any) error {
	if err := rtypecontrol.PaveArgs(args[1:], "ss"); err != nil {
		return err
	}
	rec.F = &RP{
		dns.RP{
			Mbox: args[1].(string),
			Txt:  args[2].(string),
		},
	}

	// TODO: Generate friendly Comparable and ZonefilePartial values.
	rec.Comparable = rec.F.(*RP).Mbox + " " + rec.F.(*RP).Txt
	rec.ZonefilePartial = rec.Comparable

	return nil
}

func (handle *RP) CopyToLegacyFields(rec *models.RecordConfig) {
	rp := rec.F.(*RP)
	_ = rec.SetTarget(rp.Mbox + " " + rp.Txt)
}
