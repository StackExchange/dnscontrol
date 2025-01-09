package models

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/pkg/fieldtypes"
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

// MakeFromRaw converts the RawRecordConfig into a RecordConfig by calling the
// conversion function provided when the rtype was registered.
func MakeFromRaw(typeName string, args []string, meta map[string]string, origin string) (*RecordConfig, error) {

	rt, ok := rtypeDB[typeName]
	if !ok {
		return nil, fmt.Errorf("unknown rtype %q", typeName)
	}

	rc := &RecordConfig{
		Type: typeName,
	}
	err := rt.PopulateFromRaw(rc, args, meta, origin)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

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
			rec.ImportFromLegacy(origin)
		}
	}
	return found
}

// ImportFromLegacy copies the legacy fields (MxPreference, SrvPort, etc.) to
// the .Fields structure.  It is the reverse of Seal*().
func (rc *RecordConfig) ImportFromLegacy(origin string) error {

	if IsTypeLegacy(rc.Type) {
		// Nothing to convert!
		return nil
	}

	switch rc.Type {
	case "A":
		ip, err := fieldtypes.ParseIPv4(rc.target)
		if err != nil {
			return err
		}
		return rc.PopulateFromFieldsA(ip, nil, origin)
	case "MX":
		return rc.PopulateMXFields(rc.MxPreference, rc.target, nil, origin)
	case "SRV":
		return rc.PopulateFromFieldsSRV(rc.SrvPriority, rc.SrvWeight, rc.SrvPort, rc.target, nil, origin)
	}
	panic("Should not happen")
}

// // TransformRawRecords converts the RawRecordConfigs from dnsconfig.js into RecordConfig.
// func TransformRawRecords(domains []*DomainConfig) error {

// 	for _, dc := range domains {

// 		for _, rawRec := range dc.RawRecords {

// 			rt, ok := rtypeDB[rawRec.Type]
// 			if !ok {
// 				return fmt.Errorf("unknown rtype %q", rawRec.Type)
// 			}

// 			// rc := &RecordConfig{
// 			// 	Type:      rawRec.Type,
// 			// 	SubDomain: rawRec.SubDomain,
// 			// 	Metadata:  map[string]string{},
// 			// }

// 			// Merge the metadata (convert values to string)
// 			metadata := map[string]string{}
// 			for _, m := range rawRec.Metadata {
// 				for mk, mv := range m {
// 					if v, ok := mv.(string); ok {
// 						metadata[mk] = v // Already a string
// 					} else {
// 						metadata[mk] = fmt.Sprintf("%v", mv)
// 					}
// 				}
// 			}

// 			var origin string
// 			if rawRec.SubDomain == "" {
// 				origin = dc.Name
// 			} else {
// 				origin = rawRec.SubDomain + "." + dc.Name
// 				rawRec.SubDomain = ""
// 			}

// 			rc, err := rt.MakeFromRaw(rawRec.Args, metadata, origin)
// 			if err != nil {
// 				return fmt.Errorf("%s (%q, dom=%q) record error: %w",
// 					rawRec.Type,
// 					rc.Name,
// 					dc.Name,
// 					err)
// 			}

// 			// Set the TTL
// 			if rc.TTL == 0 {
// 				ttl := rawRec.TTL
// 				if ttl == 0 {
// 					ttl = dc.DefaultTTL
// 				}
// 				rc.TTL = ttl
// 			}

// 			// Free memeory:
// 			clear(rawRec.Args)
// 			rawRec.Args = nil

// 			// Store the RecordConfig in the DomainConfig.
// 			if rawRec.EnsureAbsent {
// 				// This is a RecordConfig to be deleted.
// 				dc.EnsureAbsent = append(dc.EnsureAbsent, rc)
// 			} else {
// 				// This is a RecordConfig to be kept.
// 				dc.Records = append(dc.Records, rc)
// 			}
// 		}
// 		dc.RawRecords = nil
// 	}

// 	return nil
// }

// TransformRawRecords converts the RawRecordConfigs from dnsconfig.js into RecordConfig.
func TransformRawRecords(domains []*DomainConfig) error {

	for _, dc := range domains {

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
				return fmt.Errorf("unknown rtype %q", rawRec.Type)
			}

			err := rt.PopulateFromRaw(rec, rawRec.Args, rec.Metadata, dc.Name)
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
