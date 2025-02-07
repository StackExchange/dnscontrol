package models

import "github.com/StackExchange/dnscontrol/v4/pkg/fieldtypes"

func init() {
	MustRegisterType("A", RegisterOpts{PopulateFromRaw: PopulateFromRawA})
	MustRegisterType("MX", RegisterOpts{PopulateFromRaw: PopulateFromRawMX})
	MustRegisterType("CFSINGLEREDIRECT", RegisterOpts{PopulateFromRaw: PopulateFromRawCFSINGLEREDIRECT})
	MustRegisterType("SRV", RegisterOpts{PopulateFromRaw: PopulateFromRawSRV})

}

// RecordType is a constraint for DNS records.
type RecordType interface {
	A | MX | CFSINGLEREDIRECT | SRV
}

// CFSINGLEREDIRECT is the fields needed to store a DNS record of type CFSINGLEREDIRECT.
type CFSINGLEREDIRECT struct {
	Code             uint16 `json:"code,omitempty"`
	SRName           string `json:"sr_name,omitempty"`
	SRWhen           string `json:"sr_when,omitempty"`
	SRThen           string `json:"sr_then,omitempty"`
	SRRRulesetID     string `json:"sr_rulesetid,omitempty"`
	SRRRulesetRuleID string `json:"sr_rulesetruleid,omitempty"`
	SRDisplay        string `json:"sr_display,omitempty"`
	PRWhen           string `dns:"skip" json:"pr_when,omitempty"`
	PRThen           string `dns:"skip" json:"pr_then,omitempty"`
	PRPriority       int    `dns:"skip" json:"pr_priority,omitempty"`
	PRDisplay        string `dns:"skip" json:"pr_display,omitempty"`
}

// SRV is the fields needed to store a DNS record of type SRV.
type SRV struct {
	Priority uint16
	Weight   uint16
	Port     uint16
	Target   string `dns:"domain-name"`
}

// A is the fields needed to store a DNS record of type A.
type A struct {
	A fieldtypes.IPv4 `dns:"a"`
}

// MX is the fields needed to store a DNS record of type MX.
type MX struct {
	Preference uint16
	Mx         string `dns:"cdomain-name"`
}
