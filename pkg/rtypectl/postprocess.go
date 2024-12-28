package rtypectl

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
)

func TransformRawRecords(domains []*models.DomainConfig) error {

	for _, dc := range domains {

		for _, rawRec := range dc.RawRecords {

			// Prepare the label.
			label := rawRec.Args[0].(string) // Default to the first arg.
			if rawRec.SubDomain != "" {      // If D_EXTEND() is in use, append the subdomain.
				oldlabel := label
				if label == "@" {
					label = rawRec.SubDomain
				} else {
					label = label + "." + rawRec.SubDomain
				}
				fmt.Printf("DEBUG: subdomain=%q %q->%q\n", rawRec.SubDomain, oldlabel, label)
			}

			var labelFQDN string
			if label == "@" {
				labelFQDN = rawRec.SubDomain + "." + dc.Name
			} else {
				labelFQDN = label + "." + dc.Name
			}

			rec := &models.RecordConfig{
				Type:      rawRec.Type,
				TTL:       rawRec.TTL,
				Name:      label,
				NameFQDN:  labelFQDN,
				SubDomain: rawRec.SubDomain,
				Metadata:  map[string]string{},
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

			err := FromRaw(rec, dc.Name, rawRec.Type, rawRec.Args, rec.Metadata)
			if err != nil {
				return fmt.Errorf("%s (%q, dom=%q) record error: %w",
					rawRec.Type,
					rec.Name,
					dc.Name,
					err)
			}

			// case "CLOUDFLAREAPI_SINGLE_REDIRECT":
			// 	err = cfsingleredirect.FromRaw(rec, rawRec.Args)
			// 	rec.SetLabel("@", dc.Name)

			// default:
			// 	err = fmt.Errorf("unknown rawrec type=%q", rawRec.Type)
			// }

			// Free memeory:
			clear(rawRec.Args)
			rawRec.Args = nil

			if rawRec.EnsureAbsent {
				dc.EnsureAbsent = append(dc.EnsureAbsent, rec)
			} else {
				dc.Records = append(dc.Records, rec)
			}
		}
		dc.RawRecords = nil
	}

	return nil
}

func FromRaw(rc *models.RecordConfig, origin string, typeName string, args []any, meta map[string]string) error {

	rt, ok := rtypeDB[typeName]
	if !ok {
		fmt.Printf("DEBUG: %+v\n", rtypeDB)
		return fmt.Errorf("unknown rtype %q", typeName)
	}

	fn := rt.FromRaw
	return fn(rc, origin, args, meta)
}
