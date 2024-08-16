package models

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/providers/cloudflare/rtypes/rtypecfsingleredirect"
)

// RawRecordConfig stores the user-input from dnsconfig.js for a DNS
// Record.  This is later processed (in Go) to become a RecordConfig.
// NOTE: Only newer rtypes are processed this way.  Eventually the
// legacy types will be converted.
type RawRecordConfig struct {
	Type  string           `json:"type"`
	Args  []any            `json:"args,omitempty"`
	Metas []map[string]any `json:"metas,omitempty"`
	TTL   uint32           `json:"ttl,omitempty"`
}

func ConvertRawRecords(domains []*DomainConfig) error {

	var err error

	for _, dc := range domains {

		for _, rawRec := range dc.RawRecords {
			rec := &RecordConfig{
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

			label := rawRec.Args[0].(string)
			args := rawRec.Args[1:]
			switch rawRec.Type {

			case rtypecfsingleredirect.Name:
				rdata, error := rtypecfsingleredirect.FromRawArgs(args, label)
				if error != nil {
					return err
				}
				rec.Seal(dc.Name, label, rdata)

				//			case "MX":
				//				rdata, error := rtypemx.FromRawArgs(args)
				//				if error != nil {
				//					return err
				//				}
				//				rec.Seal(dc.Name, label, rdata)

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
		clear(dc.RawRecords)
		dc.RawRecords = nil
	}

	return nil
}
