package msdns

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/recordaudit"
)

// AuditRecords returns an error if any records are not
// supportable by this provider.
func AuditRecords(records []*models.RecordConfig) error {

	if err := recordaudit.TxtNoMultipleStrings(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-01

	if err := recordaudit.TxtNotEmpty(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-01

	if err := recordaudit.TxtNoBackticks(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-01

	if err := recordaudit.TxtNoDoubleQuotes(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-01

	if err := recordaudit.TxtNoSingleQuotes(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-01

	if err := recordaudit.TxtNoLen255(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-01

	return nil
}
