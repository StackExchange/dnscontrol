package hedns

import (
	"github.com/StackExchange/dnscontrol/v3/models"
)

// AuditRecords returns an error if any records are not
// supportable by this provider.
func AuditRecords(records []*models.RecordConfig) error {
	return nil
}
