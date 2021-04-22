package cloudns

import (
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/recordaudit"
)

// AuditRecords returns an error if any records are not
// supportable by this provider.
func AuditRecords(records []*models.RecordConfig) error {

	if err := recordaudit.TxtNoBackticks(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-01

	if err := recordaudit.TxtNotEmpty(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-01

	if err := recordaudit.TxtNoTrailingSpace(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-01

	if err := recordaudit.TxtNoDoubleQuotes(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-11

	if err := txtNoMultipleStrings(records); err != nil {
		return err
	}

	return nil
}

// ClouDNS NOT allow multiple TXT records with same name
// But allow values longer the 255
func txtNoMultipleStrings(records []*models.RecordConfig) error {
	for _, rc := range records {
		if rc.HasFormatIdenticalToTXT() { // TXT and similar:
			if len(rc.TxtStrings) > 1 {
				return fmt.Errorf("multiple strings in one txt")
			}
		}
	}
	return nil
}
