package rtypes

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/providers/cloudflare/rtypes/cfsingleredirect"
)

func PostProcess(domains []*models.DomainConfig) error {

	var err error

	for _, dc := range domains {

		for _, rawRec := range dc.RawRecords {
			rec := &models.RecordConfig{
				Type:     rawRec.Type,
				TTL:      rawRec.TTL,
				Name:     rawRec.Args[0].(string),
				Metadata: map[string]string{},
			}

			// Copy the metadata (convert everything to string)
			for _, m := range rawRec.Metas {
				for mk, mv := range m {
					if v, ok := mv.(string); ok {
						rec.Metadata[mk] = v // Already a string. No new malloc.
					} else {
						rec.Metadata[mk] = fmt.Sprintf("%v", mv)
					}
				}
			}

			// Call the proper initialize function.
			// TODO(tlim): Good candiate for an interface or a lookup table.
			switch rawRec.Type {

			case "CLOUDFLAREAPI_SINGLE_REDIRECT":
				err = cfsingleredirect.FromRaw(rec, rawRec.Args)
				rec.SetLabel("@", dc.Name)

			default:
				err = fmt.Errorf("unknown rawrec type=%q", rawRec.Type)
			}
			if err != nil {
				return fmt.Errorf("%s (%q, %q) record error: %w", rawRec.Type, rec.Name, dc.Name, err)
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
