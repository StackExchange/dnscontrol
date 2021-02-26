package cloudflare

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/recordaudit"
)

// RecordSupportAudit returns an error if any records are not
// supportable by this provider.
func RecordSupportAudit(records []*models.RecordConfig) error {
	return recordaudit.TxtTrailingSpace(records)
}
