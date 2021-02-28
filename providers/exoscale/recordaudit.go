package exoscale

import (
	"github.com/StackExchange/dnscontrol/v3/models"
)

// AuditRecordSupport returns an error if any records are not
// supportable by this provider.
func AuditRecordSupport(records []*models.RecordConfig) error {
	return nil
}
