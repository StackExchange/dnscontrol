package models

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/pkg/fieldtypes"
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

func (rc *RecordConfig) MustValidate() {
	// If A/MX/SRV, Fields should be filled in.
	_, ok := rtypeDB[rc.Type]
	if ok {
		if rc.Fields == nil {
			panic(fmt.Sprintf("RecordConfig %s %s has nil Fields", rc.Type, rc.Name))
		}
	}
}

// ImportFromLegacy copies the legacy fields (target, MxPreference,
// SrvPort, etc.) to the raw-based fields.
// The reverse of Seal*()
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
		return rc.PopulateAFields(ip, nil, origin)
	case "MX":
		return rc.PopulateMXFields(rc.MxPreference, rc.target, nil, origin)
	case "SRV":
		return rc.PopulateSRVFields(rc.SrvPriority, rc.SrvWeight, rc.SrvPort, rc.target, nil, origin)
	}
	panic("Should not happen")
	//return nil
}

func CheckAndFixImport(recs []*RecordConfig, origin string) bool {
	found := false
	for _, rec := range recs {
		// Was this created wrong?
		if IsTypeUpgraded(rec.Type) && rec.Fields == nil {
			found = true
			rec.ImportFromLegacy(origin)
		}
	}
	return found
}
