package rtypes

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
)

func PostProcess(domains []*models.DomainConfig) error {
	var err error

	for _, dc := range domains {
		for _, rawRec := range dc.RawRecords {

			// Create as much of the RecordConfig as we can now. Allow New() to fill in the reset.
			rec := &models.RecordConfig{
				Type:     rawRec.Type,
				TTL:      rawRec.TTL,
				Metadata: stringifyMetas(rawRec.Metas),
			}
			//rec.Name, rec.NameRaw, rec.NameUnicode := normalizeName(rawRec.Args[0].(string), dc.Name, dc.SubDomain)
			//rec.NameFQDN, rec.NameFQDNRaw, rec.NameFQDNUnicode := normalizeNameFQDN(rawRec.Args[0].(string), dc.Name, dc.SubDomain)
			// Name:
			// * Convert to lowercase.
			// * Convert to IDN and UNI.
			//
			// IDN: name + subdomain + domain

			rtypecontrol.Iface[rawRec.Type].FromArgs(rec, rawRec.Args)

			// rec := &models.RecordConfig{
			// 	Type:     rawRec.Type,
			// 	TTL:      rawRec.TTL,
			// 	Name:     rawRec.Args[0].(string),
			// 	Metadata: map[string]string{},
			// }

			// // Copy the metadata (convert everything to string)
			// for _, m := range rawRec.Metas {
			// 	for mk, mv := range m {
			// 		if v, ok := mv.(string); ok {
			// 			rec.Metadata[mk] = v // Already a string. No new malloc.
			// 		} else {
			// 			rec.Metadata[mk] = fmt.Sprintf("%v", mv)
			// 		}
			// 	}
			// }

			// // Call the proper initialize function.
			// // TODO(tlim): Good candidate for an interface or a lookup table.
			// switch rawRec.Type {
			// case "CLOUDFLAREAPI_SINGLE_REDIRECT":
			// 	err = cfsingleredirect.FromRaw(rec, rawRec.Args)
			// 	rec.SetLabel("@", dc.Name)

			// default:
			// 	err = fmt.Errorf("unknown rawrec type=%q", rawRec.Type)
			// }
			// if err != nil {
			// 	return fmt.Errorf("%s (%q, %q) record error: %w", rawRec.Type, rec.Name, dc.Name, err)
			// }

			// Free memeory:
			clear(rawRec.Args)
			rawRec.Args = nil

			dc.Records = append(dc.Records, rec)
		}
		dc.RawRecords = nil
	}

	return nil
}
