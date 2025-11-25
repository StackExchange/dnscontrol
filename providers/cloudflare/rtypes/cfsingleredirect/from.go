package cfsingleredirect

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
)

// // MakePageRule updates a RecordConfig to be a PAGE_RULE using PAGE_RULE data.
// func MakePageRule(rc *models.RecordConfig, priority int, code uint16, when, then string) error {
// 	if rc == nil {
// 		return errors.New("RecordConfig cannot be nil")
// 	}
// 	if when == "" || then == "" {
// 		return errors.New("when and then parameters cannot be empty")
// 	}

// 	display := mkPageRuleBlob(priority, code, when, then)

// 	rc.Type = "PAGE_RULE"
// 	rc.TTL = 1
// 	rc.CloudflareRedirect = &models.CloudflareSingleRedirectConfig{
// 		Code: code,
// 		//
// 		PRWhen:     when,
// 		PRThen:     then,
// 		PRPriority: priority,
// 		PRDisplay:  display,
// 	}
// 	return rc.SetTarget(display)
// }

// // mkPageRuleBlob creates the 1,301,when,then string used in displays.
// func mkPageRuleBlob(priority int, code uint16, when, then string) string {
// 	return fmt.Sprintf("%03d,%03d,%s,%s", priority, code, when, then)
// }

func MakeSingleRedirectFromAPI(rc *models.RecordConfig, code uint16, name, when, then string) error {
	return rtypecontrol.Func["SINGLEREDIRECT"].FromArgs(rc, []any{name, code, when, then})
}

// // MakeSingleRedirectFromAPI updatese a RecordConfig to be a SINGLEREDIRECT using data downloaded via the API.
// func MakeSingleRedirectFromAPI(rc *models.RecordConfig, code uint16, name, when, then string) error {
// 	// The target is the same as the name. It is the responsibility of the record creator to name it something diffable.
// 	target := targetFromAPIData(name, code, when, then)

// 	rc.CloudflareRedirect = &CloudflareSingleRedirectConfig{
// 		Code: code,
// 		//
// 		PRWhen:     "UNKNOWABLE",
// 		PRThen:     "UNKNOWABLE",
// 		PRPriority: 0,
// 		PRDisplay:  "UNKNOWABLE",
// 		//
// 		SRName:    name,
// 		SRWhen:    when,
// 		SRThen:    then,
// 		SRDisplay: target,
// 	}
// 	return rc.SetTarget(rc.CloudflareRedirect.SRDisplay)
// }

// // targetFromAPIData creates the display text used for a Redirect as received from Cloudflare's API.
// func targetFromAPIData(name string, code uint16, when, then string) string {
// 	return fmt.Sprintf("%s code=(%03d) when=(%s) then=(%s)",
// 		name,
// 		code,
// 		when,
// 		then,
// 	)
// }

// // makeSingleRedirectFromConvert updates a RecordConfig to be a SINGLEREDIRECT using data from a PAGE_RULE conversion.
// func makeSingleRedirectFromConvert(rc *models.RecordConfig,
// 	priority int,
// 	prWhen, prThen string,
// 	code uint16,
// 	srName, srWhen, srThen string,
// ) error {
// 	srDisplay := targetFromConverted(priority, code, prWhen, prThen, srWhen, srThen)

// 	sr := rc.CloudflareRedirect
// 	sr.Code = code

// 	sr.SRName = srName
// 	sr.SRWhen = srWhen
// 	sr.SRThen = srThen
// 	sr.SRDisplay = srDisplay

// 	return rc.SetTarget(rc.CloudflareRedirect.SRDisplay)
// }

// // targetFromConverted makes the display text used when a redirect was the result of converting a PAGE_RULE.
// func targetFromConverted(prPriority int, code uint16, prWhen, prThen, srWhen, srThen string) string {
// 	return fmt.Sprintf("%03d,%03d,%s,%s code=(%03d) when=(%s) then=(%s)", prPriority, code, prWhen, prThen, code, srWhen, srThen)
// }
