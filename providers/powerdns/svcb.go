package powerdns

import (
	"fmt"
	"strings"

	"github.com/DNSControl/dnscontrol/v4/models"
)

// contentHasPowerDNSSVCBAutoHints reports whether SVCB/HTTPS rdata contains
// PowerDNS's provider-specific automatic address hinting parameters.
func contentHasPowerDNSSVCBAutoHints(content string) bool {
	for _, field := range strings.Fields(content) {
		if field == "ipv4hint=auto" || field == "ipv6hint=auto" {
			return true
		}
	}
	return false
}

// recordHasPowerDNSSVCBAutoHints reports whether a RecordConfig is an
// SVCB/HTTPS record using PowerDNS automatic address hinting.
func recordHasPowerDNSSVCBAutoHints(rc *models.RecordConfig) bool {
	if rc.Type != "SVCB" && rc.Type != "HTTPS" {
		return false
	}
	return contentHasPowerDNSSVCBAutoHints(rc.SvcParams)
}

// powerDNSTargetCombined returns PowerDNS API content for a RecordConfig,
// preserving provider-specific SVCB/HTTPS auto hint values that miekg/dns
// cannot parse.
func powerDNSTargetCombined(rc *models.RecordConfig) string {
	if recordHasPowerDNSSVCBAutoHints(rc) {
		if rc.SvcParams == "" {
			return fmt.Sprintf("%d %s", rc.SvcPriority, rc.GetTargetField())
		}
		fmt.Printf("%d %s %s", rc.SvcPriority, rc.GetTargetField(), rc.SvcParams)
		return fmt.Sprintf("%d %s %s", rc.SvcPriority, rc.GetTargetField(), rc.SvcParams)
	}
	return rc.GetTargetCombined()
}
