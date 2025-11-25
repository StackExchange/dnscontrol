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

			// Create as much of the RecordConfig as we can now. Allow New() to fill in the reset.
			rec := &models.RecordConfig{
				Type:     rawRec.Type,
				TTL:      rawRec.TTL,
				Metadata: stringifyMetas(rawRec.Metas),
				//FilePos:  models.FixPosition(rawRec.FilePos),
			}

			setRecordNames(rec, dc, rawRec.Args[0].(string))

			// Fill in the .F/.Fields* fields.
			err := Func[rawRec.Type].FromArgs(rec, rawRec.Args)
			if err != nil {
				return err
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
		if rec.Name == "@" {
			rec.NameRaw = rec.Name
			rec.Name = rec.Name
		} else {
			rec.Name = domaintags.EfficientToASCII(n)
			rec.NameRaw = n
			rec.Name = domaintags.EfficientToUnicode(n)
		}
		rec.NameFQDN = dnsutil.AddOrigin(rec.Name, dc.Name)
		rec.NameFQDNRaw = dnsutil.AddOrigin(rec.NameRaw, dc.NameRaw)
		rec.NameFQDNUnicode = dnsutil.AddOrigin(rec.NameUnicode, dc.NameUnicode)
	} else {
		// _EXTEND() mode:
		// FIXME(tlim): Not implemented.
		if rec.Name == "@" {
			rec.NameRaw = rec.Name
			rec.Name = rec.Name
		} else {
			rec.Name = domaintags.EfficientToASCII(n)
			rec.NameRaw = n
			rec.Name = domaintags.EfficientToUnicode(n)
		}
		rec.NameFQDN = dnsutil.AddOrigin(rec.Name, dc.Name)
		rec.NameFQDNRaw = dnsutil.AddOrigin(rec.NameRaw, dc.NameRaw)
		rec.NameFQDNUnicode = dnsutil.AddOrigin(rec.NameUnicode, dc.NameUnicode)
	}
}
