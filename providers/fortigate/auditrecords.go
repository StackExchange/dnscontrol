package fortigate

import (
	"fmt"
	"github.com/StackExchange/dnscontrol/v4/models"
	"strings"
)

// AuditRecords performs basic validation and returns warnings for known limitations.
func AuditRecords(records []*models.RecordConfig) []error {
	var problems []error

	for _, rc := range records {
		switch rc.Type {
		case "A", "AAAA", "CNAME":
			// Supported – no problem.
		case "NS,PTR":
			// FortiGate limitations: these record types are not fully supported.
			problems = append(problems,
				fmt.Errorf("record type %s is not supported by FortiGate provider (name: %s)", rc.Type, rc.GetLabelFQDN()))
		default:
			problems = append(problems,
				fmt.Errorf("record type %s is not supported by FortiGate provider (name: %s)", rc.Type, rc.GetLabelFQDN()))
		}

		if rc.Type == "CNAME" && rc.GetLabel() == "@" {
			problems = append(problems,
				fmt.Errorf("CNAME at apex (@) is not allowed (name: %s)", rc.GetLabelFQDN()))
		}

		// Wildcard support
		if strings.Contains(rc.GetLabelFQDN(), "*") {
			problems = append(problems,
				fmt.Errorf("wildcard record %s is not supported by FortiGate", rc.GetLabelFQDN()))
		}
	}

	return problems
}
