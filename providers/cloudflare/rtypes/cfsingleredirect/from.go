package cfsingleredirect

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
)

// MakePageRule updates a RecordConfig to be a PAGE_RULE using PAGE_RULE data.
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
	}
	rc.SetTarget(display)
}

// mkPageRuleBlob creates the 1,301,when,then string used in displays.
func mkPageRuleBlob(priority int, code uint16, when, then string) string {
	return fmt.Sprintf("%d,%03d,%s,%s", priority, code, when, then)
}

// makeSingleRedirectFromRawRec updates a RecordConfig to be a
// SINGLEREDIRECT using the data from a RawRecord.
func makeSingleRedirectFromRawRec(rc *models.RecordConfig, code uint16, name, when, then string) {
	target := targetFromRaw(name, code, when, then)

	rc.Type = SINGLEREDIRECT
	rc.TTL = 1
	rc.CloudflareRedirect = &models.CloudflareSingleRedirectConfig{
		Code: code,
		//
		PRWhen:     "UNKNOWABLE",
		PRThen:     "UNKNOWABLE",
		PRPriority: 0,
		PRDisplay:  "UNKNOWABLE",
		//
		SRName:    name,
		SRWhen:    when,
		SRThen:    then,
		SRDisplay: target,
	}
	rc.SetTarget(rc.CloudflareRedirect.SRDisplay)
}

// targetFromRaw create the display text used for a normal Redirect.
func targetFromRaw(name string, code uint16, when, then string) string {
	return fmt.Sprintf("%s code=(%03d) when=(%s) then=(%s)",
		name,
		code,
		when,
		then,
	)
}

// MakeSingleRedirectFromAPI updatese a RecordConfig to be a SINGLEREDIRECT using data downloaded via the API.
func MakeSingleRedirectFromAPI(rc *models.RecordConfig, code uint16, name, when, then string) {
	// The target is the same as the name. It is the responsibility of the record creator to name it something diffable.
	target := targetFromAPIData(name, code, when, then)

	rc.Type = SINGLEREDIRECT
	rc.TTL = 1
	rc.CloudflareRedirect = &models.CloudflareSingleRedirectConfig{
		Code: code,
		//
		PRWhen:     "UNKNOWABLE",
		PRThen:     "UNKNOWABLE",
		PRPriority: 0,
		PRDisplay:  "UNKNOWABLE",
		//
		SRName:    name,
		SRWhen:    when,
		SRThen:    then,
		SRDisplay: target,
	}
	rc.SetTarget(rc.CloudflareRedirect.SRDisplay)
}

// targetFromAPIData creates the display text used for a Redirect as received from Cloudflare's API.
func targetFromAPIData(name string, code uint16, when, then string) string {
	return fmt.Sprintf("%s code=(%03d) when=(%s) then=(%s)",
		name,
		code,
		when,
		then,
	)
}

// makeSingleRedirectFromConvert updates a RecordConfig to be a SINGLEREDIRECT using data from a PAGE_RULE conversion.
func makeSingleRedirectFromConvert(rc *models.RecordConfig,
	priority int,
	prWhen, prThen string,
	code uint16,
	srName, srWhen, srThen string) {

	srDisplay := targetFromConverted(priority, code, prWhen, prThen, srWhen, srThen)

	rc.Type = SINGLEREDIRECT
	rc.TTL = 1
	sr := rc.CloudflareRedirect
	sr.Code = code

	sr.SRName = srName
	sr.SRWhen = srWhen
	sr.SRThen = srThen
	sr.SRDisplay = srDisplay

	rc.SetTarget(rc.CloudflareRedirect.SRDisplay)
}

// targetFromConverted makes the display text used when a redirect was the result of converting a PAGE_RULE.
func targetFromConverted(prPriority int, code uint16, prWhen, prThen, srWhen, srThen string) string {
	return fmt.Sprintf("%d,%03d,%s,%s code=(%03d) when=(%s) then=(%s)", prPriority, code, prWhen, prThen, code, srWhen, srThen)
}
