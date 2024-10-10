package cloudflare

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/providers/cloudflare/rtypes/rtypesingleredirect"
)

// MakePageRule updates a RecordConfig to be a PAGE_RULE using PAGE_RULE data.
func MakePageRule(rc *models.RecordConfig, priority int, code uint16, when, then string) {
	display := mkPageRuleBlob(priority, code, when, then)

	rc.Type = "PAGE_RULE"
	rc.TTL = 1
	rc.CloudflareRedirect = &rtypesingleredirect.SingleRedirect{
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
