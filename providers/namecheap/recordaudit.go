package namecheap

import "github.com/StackExchange/dnscontrol/v3/models"

// RecordSupportAudit returns an error if any records are not
// supportable by this provider.
func RecordSupportAudit(records []*models.RecordConfig) error {
	return nil
}


