package vultr

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/recordaudit"
)

// AuditRecords returns an error if any records are not
// supportable by this provider.
func AuditRecords(records []*models.RecordConfig) error {

	// TODO(tlim) Needs investigation. Could be a dnscontrol issue or
	// the provider doesn't support double quotes.
	if err := recordaudit.TxtNoDoubleQuotes(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-02

	return nil
}
