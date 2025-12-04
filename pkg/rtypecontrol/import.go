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

			rec, err := NewRecordConfigFromRaw(rawRec.Type, rawRec.TTL, rawRec.Args, dc)
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

func NewRecordConfigFromString(name string, ttl uint32, t string, s string, dc *models.DomainConfig) (*models.RecordConfig, error) {
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
	return NewRecordConfigFromStruct(name, ttl, t, rec, dc)

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

	err := Func[t].FromStruct(dc, rec, name, fields)
	if err != nil {
		return nil, err
	}

	return rec, nil
}

// setRecordNames updates the .Name* fields.
func setRecordNames(rec *models.RecordConfig, dc *models.DomainConfig, n string) {
	// FYI(tlim): This code could be collapse
	if rec.SubDomain == "" {
		// Not _EXTEND() mode:
		if n == "@" {
			rec.Name = "@"
			rec.NameRaw = "@"
			rec.NameUnicode = "@"
			rec.NameFQDN = dc.Name
			rec.NameFQDNRaw = dc.NameRaw
			rec.NameFQDNUnicode = dc.NameUnicode
			rec.NameFQDN = dc.Name
			rec.NameFQDNRaw = dc.NameRaw
			rec.NameFQDNUnicode = dc.NameUnicode
		} else {
			rec.Name = strings.ToLower(domaintags.EfficientToASCII(n))
			rec.NameRaw = n
			rec.NameUnicode = domaintags.EfficientToUnicode(n)
			rec.NameFQDN = rec.Name + "." + dc.Name
			rec.NameFQDNRaw = rec.NameRaw + "." + dc.NameRaw
			rec.NameFQDNUnicode = rec.NameUnicode + "." + dc.NameUnicode
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
			rec.NameFQDN = rec.Name + "." + dc.Name
			rec.NameFQDNRaw = rec.NameRaw + "." + dc.NameRaw
			rec.NameFQDNUnicode = rec.NameUnicode + "." + dc.NameUnicode
		} else {
			rec.Name = domaintags.EfficientToASCII(n) + "." + sdASCII
			rec.NameRaw = n + "." + sdRaw
			rec.NameUnicode = domaintags.EfficientToUnicode(rec.Name)
			rec.NameFQDN = rec.Name + "." + dc.Name
			rec.NameFQDNRaw = rec.NameRaw + "." + dc.NameRaw
			rec.NameFQDNUnicode = rec.NameUnicode + "." + dc.NameUnicode
		}
	}
}
