package digitalocean

import (
	"github.com/StackExchange/dnscontrol/v3/models"
)

// AuditRecords returns an error if any records are not
// supportable by this provider.
func AuditRecords(records []*models.RecordConfig) error {

	// TXT:
	// "A broken parser prevents TXTMulti strings from including double-quotes;
	// The total length of all strings can't be longer than 512; and in reality must be shorter due to sloppy validation checks."
	// "https://github.com/StackExchange/dnscontrol/issues/370"),

	// CAA:
	//		"Semicolons not supported in issue/issuewild fields.", "https://www.digitalocean.com/docs/networking/dns/how-to/create-caa-records"),

	return nil
}
