package rtypecontrol

import "github.com/StackExchange/dnscontrol/v4/models"

// FixLegacyAll populates .F to compenstate for providers that have not been
// updated to support RecordConfigV2 when creating RecordConfig.
// It is called anywhere dc.PostProcess() is called.  dc.PostProcess() can't
// call it directly because that would cause an import cycle.
func FixLegacyAll(dc *models.DomainConfig) {
	for _, rec := range dc.Records {

		FixLegacyRecord(rec)
	}
}

// FixLegacyRecord populates .F to compenstate for providers that have not been
func FixLegacyRecord(rec *models.RecordConfig) {
	// Populate .F if needed:
	// That is... If rec.F == nil and this is a "moderne" type.
	if rec.F != nil {
		return
	}
	if fixer, ok := Func[rec.Type]; ok {
		fixer.CopyFromLegacyFields(rec)
	}
}
