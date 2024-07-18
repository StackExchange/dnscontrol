package cfsingleredirect

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
)

func mkPageRuleBlob(priority int, code uint16, when, then string) string {
	return fmt.Sprintf("%d,%03d,%s,%s", priority, code, when, then)
}

func MakePageRule(rc *models.RecordConfig, priority int, code uint16, when, then string) {
	display := mkPageRuleBlob(priority, code, when, then)

	rc.Type = "PAGE_RULE"
	rc.TTL = 1
	rc.CloudflareRedirect = &models.CloudflareSingleRedirectConfig{
		Code: code,
		//
		PRWhen:     when,
		PRThen:     then,
		PRPriority: priority,
		PRDisplay:  display,
		//
		//SRName:           "UNSET",
		//SRWhen:           "UNSET",
		//SRThen:           "UNSET",
		//SRRRulesetID:     "UNSET",
		//SRRRulesetRuleID: "UNSET",
		//SRDisplay:        "UNSET",
	}
	rc.SetTarget(display)
	//printer.Printf("DEBUG: MakePageRule rc=%+v\n", rc)
	//printer.Printf("DEBUG: MakePageRule sr=%+v\n", rc.CloudflareRedirect)
}

func MakeSingleRedirectFromRawRec(rc *models.RecordConfig, code uint16, name, when, then string) {
	target := makeSingleRedirectTarget(name, code, when, then)

	rc.Type = TypeName
	rc.TTL = 1
	rc.CloudflareRedirect = &models.CloudflareSingleRedirectConfig{
		Code: code,
		//
		PRWhen:     "UNKNOWABLE",
		PRThen:     "UNKNOWABLE",
		PRPriority: 0,
		PRDisplay:  "UNKNOWABLE",
		//
		SRName: name,
		SRWhen: when,
		SRThen: then,
		//SRRRulesetID:     "UNSET",
		//SRRRulesetRuleID: "UNSET",
		SRDisplay: target,
	}
	rc.SetTarget(rc.CloudflareRedirect.SRDisplay)
}

func MakeSingleRedirectFromAPI(rc *models.RecordConfig, code uint16, name, when, then string) {
	// The target is the same as the name. It is the responsibility of the record creator to name it something diffable.
	target := mkTargetAPI(name, code, when, then)

	rc.Type = TypeName
	rc.TTL = 1
	rc.CloudflareRedirect = &models.CloudflareSingleRedirectConfig{
		Code: code,
		//
		PRWhen:     "UNKNOWABLE",
		PRThen:     "UNKNOWABLE",
		PRPriority: 0,
		PRDisplay:  "UNKNOWABLE",
		//
		SRName: name,
		SRWhen: when,
		SRThen: then,
		//SRRRulesetID:     "UNSET",
		//SRRRulesetRuleID: "UNSET",
		SRDisplay: target,
	}
	rc.SetTarget(rc.CloudflareRedirect.SRDisplay)
}

func MakeSingleRedirectFromConvert(rc *models.RecordConfig,
	priority int,
	prWhen, prThen string,
	code uint16,
	srName, srWhen, srThen string) {

	srDisplay := mkDisplayConverted(priority, code, prWhen, prThen, srWhen, srThen)

	rc.Type = TypeName
	rc.TTL = 1
	sr := rc.CloudflareRedirect
	sr.Code = code
	//
	//PRWhen:     "UNKNOWABLE",
	//PRThen:     "UNKNOWABLE",
	//PRPriority: 0,
	//PRDisplay:  "UNKNOWABLE",
	//
	sr.SRName = srName
	sr.SRWhen = srWhen
	sr.SRThen = srThen
	//SRRRulesetID:     "UNSET",
	//SRRRulesetRuleID: "UNSET",
	sr.SRDisplay = srDisplay

	rc.SetTarget(rc.CloudflareRedirect.SRDisplay)
}

func mkDisplayConverted(prPriority int, code uint16, prWhen, prThen, srWhen, srThen string) string {
	return fmt.Sprintf("%d,%03d,%s,%s code=(%03d) when=(%s) then=(%s)", prPriority, code, prWhen, prThen, code, srWhen, srThen)
}
