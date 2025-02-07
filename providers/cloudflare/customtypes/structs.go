package cloudflarecustomtypes

import "github.com/miekg/dns"

// DO NOT USE THESE STRUCTS!

// Instead use: models.CFSINGLEREDIRECT

// go generate will copy this struct to $git/dnscontrol/models/types-cloudflareapi.go
// do deal with the fact that you can't overload models.RecordConfig
// from outside the models directory.

type CFSINGLEREDIRECT struct {
	Hdr dns.RR_Header

	//
	Code uint16 `json:"code,omitempty"` // 301 or 302
	//
	// SR == SingleRedirect
	SRName           string `json:"sr_name,omitempty"` // How is this displayed to the user
	SRWhen           string `json:"sr_when,omitempty"`
	SRThen           string `json:"sr_then,omitempty"`
	SRRRulesetID     string `json:"sr_rulesetid,omitempty"`
	SRRRulesetRuleID string `json:"sr_rulesetruleid,omitempty"`
	SRDisplay        string `json:"sr_display,omitempty"` // How is this displayed to the user (SetTarget) for CF_SINGLE_REDIRECT
	//
	// PR == PageRule
	PRWhen     string `dns:"skip" json:"pr_when,omitempty"`
	PRThen     string `dns:"skip" json:"pr_then,omitempty"`
	PRPriority int    `dns:"skip" json:"pr_priority,omitempty"` // Really an identifier for the rule.
	PRDisplay  string `dns:"skip" json:"pr_display,omitempty"`  // How is this displayed to the user (SetTarget) for CF_REDIRECT/CF_TEMP_REDIRECT
}
