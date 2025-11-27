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

			rec, err := NewRecordConfigFromRaw(rawRec.Type, rawRec.Args, dc)
			if err != nil {
				return fmt.Errorf("%s: %w", nil, err)
				// TODO(tlim): Fix FilePos
				//return fmt.Errorf("%s: %w", rawRec.FilePos, err)
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

func NewRecordConfigFromRaw(t string, args []any, dc *models.DomainConfig) (*models.RecordConfig, error) {
	fmt.Printf("DEBUG: NewRecordConfigFromRaw t=%q args=%+v\n", t, args)
	if _, ok := Func[t]; !ok {
		return nil, fmt.Errorf("record type %q is not supported", t)
	}

	// Create as much of the RecordConfig as we can now. Allow New() to fill in the reset.
	rec := &models.RecordConfig{
		Type:     t,
		Name:     args[0].(string), // May be fixed later.
		Metadata: map[string]string{},
		//FilePos:  models.FixPosition(filePos),
	}

	setRecordNames(rec, dc, args[0].(string))

	if rec.Type == "" {
		panic("rtypecontrol: NewRecordConfigFromRaw: empty record type")
	}

	// Fill in the .F/.Fields* fields.
	err := Func[t].FromArgs(dc, rec, args)
	if err != nil {
		return nil, err
	}

	return rec, nil
}

func stringifyMetas(metas []map[string]any) map[string]string {
	result := make(map[string]string)
	for _, m := range metas {
		for mk, mv := range m {
			if v, ok := mv.(string); ok {
				result[mk] = v // Already a string. No new malloc.
			} else {
				result[mk] = fmt.Sprintf("%v", mv)
			}
		}
	}
	return result
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
		rec.NameFQDN = dnsutil.AddOrigin(rec.Name, dc.Name)
		rec.NameFQDNRaw = dnsutil.AddOrigin(rec.NameRaw, dc.NameRaw)
		rec.NameFQDNUnicode = dnsutil.AddOrigin(rec.NameUnicode, dc.NameUnicode)
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
