package namedotcom

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/recordaudit"
)

// AuditRecordSupport returns an error if any records are not
// supportable by this provider.
func AuditRecordSupport(records []*models.RecordConfig) error {
	return recordaudit.TxtTrailingSpace(records)
}
