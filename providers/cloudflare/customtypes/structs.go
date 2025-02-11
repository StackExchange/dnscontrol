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
	// SR == SingleRedirect
	SRName string `json:"sr_name,omitempty"` // How is this displayed to the user
	//
	Code uint16 `json:"code,omitempty" dnscontrol:"_,redirectcode"` // 301 or 302
	//
	SRWhen           string `json:"sr_when,omitempty"`
	SRThen           string `json:"sr_then,omitempty"`
	SRRRulesetID     string `json:"sr_rulesetid,omitempty" dnscontrol:"_,noraw,noparsereturn"`
	SRRRulesetRuleID string `json:"sr_rulesetruleid,omitempty" dnscontrol:"_,noraw,noparsereturn"`
	// How is this displayed to the user (SetTarget) for CF_SINGLE_REDIRECT
	SRDisplay string `json:"sr_display,omitempty" dnscontrol:"_,srdisplay,noraw,noparsereturn"`
	//
	// PR == PageRule
	PRWhen string `json:"pr_when,omitempty" dnscontrol:"_,noraw,parsereturnunknowable"`
	PRThen string `json:"pr_then,omitempty" dnscontrol:"_,noraw,parsereturnunknowable"`

	// An identifier for the rule.
	PRPriority int `json:"pr_priority,omitempty" dnscontrol:"_,noraw,noparsereturn"`

	// How is this displayed to the user (SetTarget) for CF_REDIRECT/CF_TEMP_REDIRECT
	PRDisplay string `json:"pr_display" dnscontrol:"_,noraw,parsereturnunknowable,noparsereturn"`
}
