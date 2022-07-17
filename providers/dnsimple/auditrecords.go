package dnsimple

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/recordaudit"
)

// AuditRecords returns an error if any records are not
// supportable by this provider.
func AuditRecords(records []*models.RecordConfig) error {
	if err := recordaudit.TxtNoDoubleQuotes(records); err != nil {
		return err
	}
	return nil
}
