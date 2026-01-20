package rtype

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/domaintags"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
	"github.com/miekg/dns"
)

func init() {
	rtypecontrol.Register(&DS{})
}

// DS RR.
type DS struct {
	dns.DS
}

// Name returns the DNS record type as a string.
func (handle *DS) Name() string {
	return "DS"
}

// FromArgs fills in the RecordConfig from []any, which is typically from a parsed config file.
func (handle *DS) FromArgs(dcn *domaintags.DomainNameVarieties, rec *models.RecordConfig, args []any) error {
	if err := rtypecontrol.PaveArgs(args[1:], "WBs"); err != nil {
		return fmt.Errorf("ERROR: (%s) [DS(%q, %v)]: %w",
			rec.FilePos,
			rec.Name, rtypecontrol.StringifyQuoted(args[1:]),
			err)
	}
	fields := &DS{
		dns.DS{
			//Mbox: dnsutil.AddOrigin(args[1].(string), dcn.NameASCII+"."),
			//Txt:  dnsutil.AddOrigin(args[2].(string), dcn.NameASCII+"."),
			DsKeyTag:     keytag,
			DsAlgorithm:  algorithm,
			DsDigestType: digesttype,
			DsDigest:     digest,
		},
	}

	return handle.FromStruct(dcn, rec, args[0].(string), fields)
}

// FromStruct fills in the RecordConfig from a struct, typically from an API response.
func (handle *DS) FromStruct(dcn *domaintags.DomainNameVarieties, rec *models.RecordConfig, name string, fields any) error {
	rec.F = fields

	rec.ZonefilePartial = rec.GetTargetRFC1035Quoted()
	rec.Comparable = rec.ZonefilePartial

	return nil
}

// CopyToLegacyFields populates the legacy fields of the RecordConfig using the fields in .F.
func (handle *DS) CopyToLegacyFields(rec *models.RecordConfig) {
	ds := rec.F.(*DS)
	_ = rec.SetTarget(DS.Mbox + " " + DS.Txt)
	/* TODO: Finish this */
}
