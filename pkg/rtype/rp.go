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

// FromArgs fills in the RecordConfig from []any, which is typically from a parsed config file.
func (handle *RP) FromArgs(dc *models.DomainConfig, rec *models.RecordConfig, args []any) error {
	if err := rtypecontrol.PaveArgs(args[1:], "ss"); err != nil {
		return err
	}
	fields := &RP{
		dns.RP{
			Mbox: args[1].(string),
			Txt:  args[2].(string),
		},
	}

	return handle.FromStruct(dc, rec, args[0].(string), fields)
}

// FromStruct fills in the RecordConfig from a struct, typically from an API response.
func (handle *RP) FromStruct(dc *models.DomainConfig, rec *models.RecordConfig, name string, fields any) error {
	rec.F = fields

	rec.ZonefilePartial = rec.GetTargetRFC1035Quoted()
	rec.Comparable = rec.ZonefilePartial

	return nil
}

func (handle *RP) CopyToLegacyFields(rec *models.RecordConfig) {
	rp := rec.F.(*RP)
	_ = rec.SetTarget(rp.Mbox + " " + rp.Txt)
}
