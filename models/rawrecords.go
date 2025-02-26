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
						rec.Metadata[mk] = v // Already a string
					} else {
						rec.Metadata[mk] = fmt.Sprintf("%v", mv)
					}
				}
			}

			rt, ok := rtypeDB[rawRec.Type]
			if !ok {
				return fmt.Errorf("unknown (TRR) rtype %q", rawRec.Type)
			}

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
