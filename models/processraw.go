package models

import (
	"fmt"
)

func TransformRawRecords(domains []*DomainConfig) error {

	for _, dc := range domains {
		//fmt.Printf("DEBUG: dc.DefaultTTL = %d\n", dc.DefaultTTL)

		for _, rawRec := range dc.RawRecords {

			if rawRec.TTL == 0 {
				rawRec.TTL = dc.DefaultTTL
			}

			rec := &RecordConfig{
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

			rt, ok := rtypeDB[rawRec.Type]
			if !ok {
				return fmt.Errorf("unknown rtype %q", rawRec.Type)
			}

			err := rt.FromRaw(rec, rawRec.Args, rec.Metadata, dc.Name)
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

func FromRaw(rc *RecordConfig, origin string, typeName string, args []string, meta map[string]string) error {

	rt, ok := rtypeDB[typeName]
	if !ok {
		//fmt.Printf("DEBUG: %+v\n", rtypeDB)
		return fmt.Errorf("unknown rtype %q", typeName)
	}

	fn := rt.FromRaw
	return fn(rc, args, meta, origin)
}

// type FromRawFn func(rawfields []string, metadata map[string]string, origin string) error
