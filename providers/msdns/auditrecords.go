package msdns

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/recordaudit"
)

// AuditRecords returns an error if any records are not
// supportable by this provider.
func AuditRecords(records []*models.RecordConfig) error {

	for i, rc := range records {
		fmt.Printf("DEBUG %02d len(txts) = %d\n", i, len(rc.TxtStrings))

	}
	if err := recordaudit.TxtNoMultipleStrings(records); err != nil {
		return err
	}

	if err := recordaudit.TxtNotEmpty(records); err != nil {
		return err
	}

	// TODO(tlim): Should be easy to implement support for this.
	if err := recordaudit.TxtNoBackticks(records); err != nil {
		return err
	}

	if err := recordaudit.TxtNoDoubleQuotes(records); err != nil {
		return err
	}

	if err := recordaudit.TxtNoSingleQuotes(records); err != nil {
		return err
	}

	if err := recordaudit.TxtNoLen255(records); err != nil {
		return err
	}

	return nil
}
