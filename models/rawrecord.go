package models

// RawRecordConfig stores the user-input from dnsconfig.js for a DNS
// Record.  This is later processed (in Go) to become a RecordConfig.
// NOTE: Only newer rtypes are processed this way.  Eventually the
// legacy types will be converted.
type RawRecordConfig struct {
	Type      string           `json:"type"`
	Args      []string         `json:"args,omitempty"`
	Metas     []map[string]any `json:"metas,omitempty"`
	TTL       uint32           `json:"ttl,omitempty"`
	SubDomain string           `json:"subdomain,omitempty"`

	// Override NO_PURGE and delete this record
	EnsureAbsent bool `json:"ensure_absent,omitempty"`
}
