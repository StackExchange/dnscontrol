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
	if err := rtypecontrol.PaveArgs(args[1:], "wbbs"); err != nil {
		return fmt.Errorf("ERROR: (%s) [DS(%q, %v)]: %w",
			rec.FilePos,
			rec.Name, rtypecontrol.StringifyQuoted(args[1:]),
			err)
	}
	fields := &dns.DS{
		KeyTag:     args[1].(uint16),
		Algorithm:  args[2].(uint8),
		DigestType: args[3].(uint8),
		Digest:     args[4].(string),
	}
	//fmt.Printf("DEBUG: DS.FromArgs: populated fields: %+v\n", fields)
	//fmt.Printf("DEBUG: DS.FromArgs: populated fields: %T\n", fields)

	return handle.FromStruct(dcn, rec, args[0].(string), fields)
}

// FromStruct fills in the RecordConfig from a struct, typically from an API response.
func (handle *DS) FromStruct(dcn *domaintags.DomainNameVarieties, rec *models.RecordConfig, name string, fields any) error {
	// Since we accept "any", we must verify the type. It should be the "inner" type of .F, not the rtype.DS{}.
	// These bugs will be caught during integration testing.
	ds, ok := fields.(*dns.DS)
	if !ok {
		return fmt.Errorf("fields is not *dns.DS, got %T", fields)
	}
	rec.F = &DS{*ds}

	rec.ZonefilePartial = rec.GetTargetRFC1035Quoted()
	rec.Comparable = rec.ZonefilePartial

	handle.CopyToLegacyFields(rec)
	return nil
}

// CopyToLegacyFields populates the legacy fields of the RecordConfig using the fields in .F.
func (handle *DS) CopyToLegacyFields(rec *models.RecordConfig) {
	// fmt.Printf("DEBUG: DS.CopyTo: : rec   %T\n", rec)
	// fmt.Printf("DEBUG: DS.CopyTo: : rec.F %T\n", rec.F)
	// x := rec.F
	// fmt.Printf("DEBUG: DS.CopyTo: : x     %T\n", x)

	ds := rec.F.(*DS)
	_ = rec.SetTargetDS(ds.KeyTag, ds.Algorithm, ds.DigestType, ds.Digest)
}

// CopyFromLegacyFields uses the the legacy fields to populate .F
func (handle *DS) CopyFromLegacyFields(rec *models.RecordConfig) {
	rec.F = &DS{
		dns.DS{
			KeyTag:     rec.DsKeyTag,
			Algorithm:  rec.DsAlgorithm,
			DigestType: rec.DsDigestType,
			Digest:     rec.DsDigest,
		},
	}
}
