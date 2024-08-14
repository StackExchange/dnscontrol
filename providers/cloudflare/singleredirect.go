package cloudflare

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/providers/cloudflare/rtypes/rtypesingleredirect"
)

// MakeSingleRedirectFromAPI updatese a RecordConfig to be a SINGLEREDIRECT using data downloaded via the API.
func MakeSingleRedirectFromAPI(rc *models.RecordConfig, code uint16, name, when, then string) {
	// The target is the same as the name. It is the responsibility of the record creator to name it something diffable.
	target := targetFromAPIData(name, code, when, then)

	rc.Type = rtypesingleredirect.Name
	rc.TTL = 1
	rc.Rdata = &rtypesingleredirect.SingleRedirect{
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
	rc.SetTarget(rc.AsSingleRedirect().SRDisplay)
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
