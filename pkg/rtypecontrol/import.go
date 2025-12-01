package rtypecontrol

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/domaintags"
	"github.com/miekg/dns/dnsutil"
)

// ImportRawRecords imports the RawRecordConfigs into RecordConfigs.
func ImportRawRecords(domains []*models.DomainConfig) error {

	for _, dc := range domains {
		for _, rawRec := range dc.RawRecords {

			rec, err := NewRecordConfigFromRaw(rawRec.Type, rawRec.TTL, rawRec.Args, dc)
			rec.FilePos = models.FixPosition(rawRec.FilePos)
			if err != nil {
				return fmt.Errorf("%s: %w", rec.FilePos, err)
			}

			// Free memeory:
			clear(rawRec.Args)
			rawRec.Args = nil

			dc.Records = append(dc.Records, rec)
		}
		dc.RawRecords = nil
	}

	return nil
}

func NewRecordConfigFromRaw(t string, ttl uint32, args []any, dc *models.DomainConfig) (*models.RecordConfig, error) {
	if _, ok := Func[t]; !ok {
		return nil, fmt.Errorf("record type %q is not supported", t)
	}
	if t == "" {
		panic("rtypecontrol: NewRecordConfigFromRaw: empty record type")
	}

	// Create as much of the RecordConfig as we can now. Allow New() to fill in the reset.
	rec := &models.RecordConfig{
		Type:     t,
		TTL:      ttl,
		Metadata: map[string]string{},
	}
	setRecordNames(rec, dc, args[0].(string))

	// Fill in the .F/.Fields* fields.
	err := Func[t].FromArgs(dc, rec, args)
	if err != nil {
		return nil, err
	}

	return rec, nil
}

func NewRecordConfigFromStruct(name string, ttl uint32, t string, fields any, dc *models.DomainConfig) (*models.RecordConfig, error) {
	if _, ok := Func[t]; !ok {
		return nil, fmt.Errorf("record type %q is not supported", t)
	}
	if t == "" {
		panic("rtypecontrol: NewRecordConfigFromStruct: empty record type")
	}

	// Create as much of the RecordConfig as we can now. Allow New() to fill in the reset.
	rec := &models.RecordConfig{
		Type:     t,
		TTL:      ttl,
		Metadata: map[string]string{},
	}
	setRecordNames(rec, dc, name)

	// // Fill in the .F/.Fields* fields.
	// err := Func[t].FromArgs(dc, rec, []any{name, fields.(*dns.RP).Mbox, fields.(*dns.RP).Txt})
	// if err != nil {
	// 	return nil, err
	// }
	err := Func[t].FromStruct(dc, rec, name, fields)
	if err != nil {
		return nil, err
	}

	return rec, nil
}

func setRecordNames(rec *models.RecordConfig, dc *models.DomainConfig, n string) {

	if rec.SubDomain == "" {
		// Not _EXTEND() mode:
		if n == "@" {
			rec.Name = "@"
			rec.NameRaw = "@"
			rec.NameUnicode = "@"
		} else {
			rec.Name = domaintags.EfficientToASCII(n)
			rec.NameRaw = n
			rec.NameUnicode = domaintags.EfficientToUnicode(n)
		}
		rec.NameFQDN = dc.Name
		rec.NameFQDNRaw = dc.NameRaw
		rec.NameFQDNUnicode = dc.NameUnicode
	} else {
		// _EXTEND() mode:
		// FIXME(tlim): Not implemented.
		sdRaw := rec.SubDomain
		sdIDN := domaintags.EfficientToASCII(rec.SubDomain)
		sdUnicode := domaintags.EfficientToUnicode(rec.SubDomain)
		if n == "@" {
			rec.Name = sdIDN
			rec.NameRaw = sdRaw
			rec.NameUnicode = sdUnicode
		} else {
			rec.Name = domaintags.EfficientToASCII(n + "." + sdIDN)
			rec.NameRaw = n + "." + sdRaw
			rec.NameUnicode = domaintags.EfficientToUnicode(n + "." + sdUnicode)
		}
		rec.NameFQDN = dnsutil.AddOrigin(rec.Name, dc.Name)
		rec.NameFQDNRaw = dnsutil.AddOrigin(rec.NameRaw, dc.NameRaw)
		rec.NameFQDNUnicode = dnsutil.AddOrigin(rec.NameUnicode, dc.NameUnicode)
	}
}
