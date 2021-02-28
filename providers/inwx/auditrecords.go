package inwx

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/recordaudit"
)

// AuditRecords returns an error if any records are not
// supportable by this provider.
func AuditRecords(records []*models.RecordConfig) error {

	// Bug in the API prevents these from working.
	if err := recordaudit.TxtNoBackticks(records); err != nil {
		return err
	}

	if err := recordaudit.TxtNotEmpty(records); err != nil {
		return err
	}

	if err := recordaudit.TxtNoLen255(records); err != nil {
		return err
	}

	if err := recordaudit.TxtNoTrailingSpace(records); err != nil {
		return err
	}

	return nil
}
