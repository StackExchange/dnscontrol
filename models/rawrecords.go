package models

import (
	"fmt"

	"github.com/go-acme/lego/v4/log"
)

// RawRecordConfig stores the raw user-input from dnsconfig.js for a DNS Record
// (A, MX, SRV, etc).  This is later processed to become a RecordConfig.  NOTE:
// Only newer rtypes are processed this way.  Eventually the legacy types will
// be converted and removed.
type RawRecordConfig struct {
	Type      string           `json:"type"`
	Args      []string         `json:"args,omitempty"`
	Metadata  []map[string]any `json:"metas,omitempty"`
	TTL       uint32           `json:"ttl,omitempty"`
	SubDomain string           `json:"subdomain,omitempty"`

	// Override NO_PURGE and delete this record
	EnsureAbsent bool `json:"ensure_absent,omitempty"`
}

// // FromRaw converts the RawRecordConfig into a RecordConfig by calling the
// // conversion function provided when the rtype was registered.
// func FromRaw(rc *RecordConfig, origin string, typeName string, args []string, meta map[string]string) error {

// 	rt, ok := rtypeDB[typeName]
// 	if !ok {
// 		return fmt.Errorf("unknown (FromRaw) rtype %q", typeName)
// 	}

// 	return rt.PopulateFromRaw(rc, args, meta, effectiveOrigin(rc.SubDomain, origin))
// }

// CheckAndFixImport checks the records for any that were created with a
// provider that has not yet been upgraded. In theory leaving providers in the
// legacy state should not cause any issues, but it is a good idea to fix them
// as soon as possible.
func CheckAndFixImport(recs []*RecordConfig, origin string) bool {
	found := false
	for _, rec := range recs {
		//fmt.Printf("DEBUG: Found record %s %s %v\n", rec.Type, rec.Name, rec)
		// Was this created wrong?
		if IsTypeUpgraded(rec.Type) && rec.Fields == nil {
			found = true
			log.Warnf("LEGACY PROVIDER needs fixing! Created invalid record: %s %s %v\n", rec.Type, rec.Name, rec)
			if err := rec.ImportFromLegacy(origin); err != nil {
				log.Warnf("Error fixing record: %s %s %v: %v\n", rec.Type, rec.Name, rec, err)
			}
		}
	}
	return found
}

// MustImportFromLegacy is like ImportFromLegacy but panics on error. Use only in tests and init() functions.
func (rc *RecordConfig) MustImportFromLegacy(origin string) {
	if err := rc.ImportFromLegacy(origin); err != nil {
		panic(err)
	}
}

// TransformRawRecords converts the RawRecordConfigs from dnsconfig.js into RecordConfig.
func TransformRawRecords(domains []*DomainConfig) error {

	for _, dc := range domains {

		// fmt.Printf("DEBUG: TransformRawRecords: rawRecords=%+v\n", dc.RawRecords)

		for _, rawRec := range dc.RawRecords {
			// fmt.Printf("DEBUG: TransformRawRecords: record=%+v\n", rawRec)

			if rawRec.TTL == 0 {
				rawRec.TTL = dc.DefaultTTL
			}

			rec := &RecordConfig{
				Type:     rawRec.Type,
				TTL:      rawRec.TTL,
				Metadata: map[string]string{},
			}
			subdomain := rawRec.SubDomain
			// fmt.Printf("DEBUG: TransformRawRecords: subdomain=%v\n", subdomain)

			// Copy the metadata (convert values to string)
			//fmt.Printf("DEBUG: TransformRawRecords: %v\n", rawRec.Metadata)
			for _, m := range rawRec.Metadata {
				for mk, mv := range m {
					if v, ok := mv.(string); ok {
						//fmt.Printf("DEBUG: TransformRawRecords: meta add: %q : %q\n", mk, v)
						rec.Metadata[mk] = v // Already a string
					} else {
						//fmt.Printf("DEBUG: TransformRawRecords: meta add: %q : %q\n", mk, mv)
						rec.Metadata[mk] = fmt.Sprintf("%v", mv)
					}
				}
			}

			rt, ok := rtypeDB[rawRec.Type]
			if !ok {
				return fmt.Errorf("unknown (TRR) rtype %q", rawRec.Type)
			}

			// fmt.Printf("DEBUG: TransformRawRecords: rec=%v subdomain=%v args=%v\n", rec, subdomain, rawRec.Args)
			err := rt.PopulateFromRaw(rec, rawRec.Args, rec.Metadata, subdomain, dc.Name)
			if err != nil {
				return fmt.Errorf("%s (label=%q, zone=%q args=%v) record error: %w",
					rawRec.Type,
					rec.Name,
					dc.Name,
					rawRec.Args,
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

// // effectiveName returns the effective origin given a "subdomain" and an
// // "origin".  The concept of a subdomain is only relevant in dnsconfig.js and
// // RawRecordConfig.  In the RecordConfig, the "Name" field is the full name
// // (minor the dc.Name) and any .Target or other fields are FQDNs or relative to
// // the effective origin.
// func effectiveTarget(sub, origin string) string {
// 	fmt.Printf("DEBUG: effectiveOrigin: %q %q\n", sub, origin)
// 	if sub == "" {
// 		fmt.Printf("DEBUG: effectiveOrigin: result=%q\n", origin)
// 		return origin
// 	}
// 	x := sub + "." + origin
// 	fmt.Printf("DEBUG: effectiveOrigin: result=%q\n", x)
// 	return x
// }

// func effectiveLabel(short, sub string) string {
// 	fmt.Printf("DEBUG: effectiveLabel: called %q %q\n", short, sub)
// 	var result string

// 	if sub == "" {
// 		// Not in D_EXTEND() mode.
// 		result = short
// 	} else if short == "" || short == "@" {
// 		// In D_EXTEND() mode.  Short is the (fake) origin.
// 		result = sub
// 	} else {
// 		// In D_EXTEND() mode.
// 		result = short + "." + sub
// 	}
// 	fmt.Printf("DEBUG: effectiveLabel: returned %q\n", result)
// 	return result
// }
