package inwx

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/recordaudit"
)

// AuditRecords returns an error if any records are not
// supportable by this provider.
func AuditRecords(records []*models.RecordConfig) error {
	var err error

	// Bug in the API prevents these from working.
	err = recordaudit.TxtBackticks(records)
	if err != nil {
		return err
	}

	err = recordaudit.TxtEmpty(records)
	if err != nil {
		return err
	}

	err = recordaudit.TxtLen255(records)
	if err != nil {
		return err
	}

	err = recordaudit.TxtTrailingSpace(records)
	if err != nil {
		return err
	}

	return nil
}
