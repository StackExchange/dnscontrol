package rtypectl

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
)

func TransformRawRecords(domains []*models.DomainConfig) error {

	for _, dc := range domains {
		//fmt.Printf("DEBUG: dc.DefaultTTL = %d\n", dc.DefaultTTL)

		for _, rawRec := range dc.RawRecords {

			if rawRec.TTL == 0 {
				rawRec.TTL = dc.DefaultTTL
			}

			rec := &models.RecordConfig{
				Type:      rawRec.Type,
				TTL:       rawRec.TTL,
				SubDomain: rawRec.SubDomain,
				Metadata:  map[string]string{},
			}

			// Copy the metadata (convert values to string)
			for _, m := range rawRec.Metas {
				for mk, mv := range m {
					if v, ok := mv.(string); ok {
						rec.Metadata[mk] = v // Already a string
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

func FromRaw(rc *models.RecordConfig, origin string, typeName string, args []string, meta map[string]string) error {

	rt, ok := rtypeDB[typeName]
	if !ok {
		//fmt.Printf("DEBUG: %+v\n", rtypeDB)
		return fmt.Errorf("unknown rtype %q", typeName)
	}

	fn := rt.FromRaw
	return fn(rc, origin, args, meta)
}
