package rtypecontrol

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/domaintags"
	"github.com/miekg/dns"
)

// ImportRawRecords imports the RawRecordConfigs into RecordConfigs.
func ImportRawRecords(domains []*models.DomainConfig) error {
	for _, dc := range domains {
		for _, rawRec := range dc.RawRecords {

			rec, err := NewRecordConfigFromRaw(rawRec.Type, rawRec.TTL, rawRec.Args, dc.DomainNameVarieties())
			if err != nil {
				return err
			}
			rec.FilePos = models.FixPosition(rawRec.FilePos)

			// Free memeory:
			clear(rawRec.Args)
			rawRec.Args = nil

			dc.Records = append(dc.Records, rec)
		}
		dc.RawRecords = nil
	}

	return nil
}

func NewRecordConfigFromRaw(t string, ttl uint32, args []any, dcn *domaintags.DomainNameVarieties) (*models.RecordConfig, error) {
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
	setRecordNames(rec, dcn, args[0].(string))

	// Fill in the .F/.Fields* fields.
	err := Func[t].FromArgs(dcn, rec, args)
	if err != nil {
		return nil, err
	}

	return rec, nil
}

func NewRecordConfigFromString(name string, ttl uint32, t string, s string, dcn *domaintags.DomainNameVarieties) (*models.RecordConfig, error) {
	if _, ok := Func[t]; !ok {
		return nil, fmt.Errorf("record type %q is not supported", t)
	}
	if t == "" {
		panic("rtypecontrol: NewRecordConfigFromStruct: empty record type")
	}

	rec, err := dns.NewRR(fmt.Sprintf("$ORIGIN .\n. %d IN %s %s", ttl, t, s))
	if err != nil {
		return nil, err
	}
	return NewRecordConfigFromStruct(name, ttl, t, rec, dcn)

}

func NewRecordConfigFromStruct(name string, ttl uint32, t string, fields any, dcn *domaintags.DomainNameVarieties) (*models.RecordConfig, error) {
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
	setRecordNames(rec, dcn, name)

	err := Func[t].FromStruct(dcn, rec, name, fields)
	if err != nil {
		return nil, err
	}

	return rec, nil
}

// setRecordNames updates the .Name* fields.
func setRecordNames(rec *models.RecordConfig, dcn *domaintags.DomainNameVarieties, n string) {
	// FYI(tlim): This code could be collapse
	if rec.SubDomain == "" {
		// Not _EXTEND() mode:
		if n == "@" {
			rec.Name = "@"
			rec.NameRaw = "@"
			rec.NameUnicode = "@"
			rec.NameFQDN = dcn.NameASCII
			rec.NameFQDNRaw = dcn.NameRaw
			rec.NameFQDNUnicode = dcn.NameUnicode
			rec.NameFQDN = dcn.NameASCII
			rec.NameFQDNRaw = dcn.NameRaw
			rec.NameFQDNUnicode = dcn.NameUnicode
		} else {
			rec.Name = strings.ToLower(domaintags.EfficientToASCII(n))
			rec.NameRaw = n
			rec.NameUnicode = domaintags.EfficientToUnicode(n)
			rec.NameFQDN = rec.Name + "." + dcn.NameASCII
			rec.NameFQDNRaw = rec.NameRaw + "." + dcn.NameRaw
			rec.NameFQDNUnicode = rec.NameUnicode + "." + dcn.NameUnicode
		}
	} else {
		// D_EXTEND() mode:
		sdRaw := rec.SubDomain
		sdASCII := strings.ToLower(domaintags.EfficientToASCII(rec.SubDomain))
		sdUnicode := domaintags.EfficientToUnicode(sdASCII)
		if n == "@" {
			rec.Name = sdASCII
			rec.NameRaw = sdRaw
			rec.NameUnicode = sdUnicode
			rec.NameFQDN = rec.Name + "." + dcn.NameASCII
			rec.NameFQDNRaw = rec.NameRaw + "." + dcn.NameRaw
			rec.NameFQDNUnicode = rec.NameUnicode + "." + dcn.NameUnicode
		} else {
			rec.Name = domaintags.EfficientToASCII(n) + "." + sdASCII
			rec.NameRaw = n + "." + sdRaw
			rec.NameUnicode = domaintags.EfficientToUnicode(rec.Name)
			rec.NameFQDN = rec.Name + "." + dcn.NameASCII
			rec.NameFQDNRaw = rec.NameRaw + "." + dcn.NameRaw
			rec.NameFQDNUnicode = rec.NameUnicode + "." + dcn.NameUnicode
		}
	}
}
