package dnsimple

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/recordaudit"
)

// AuditRecords returns an error if any records are not
// supportable by this provider.
func AuditRecords(records []*models.RecordConfig) error {
	//TODO(onlyhavecans) I think we can support multiple strings.
	if err := recordaudit.TxtNoMultipleStrings(records); err != nil {
		return err
	}

	if err := recordaudit.TxtNoTrailingSpace(records); err != nil {
		return err
	} // as of 2022-07

	if err := recordaudit.TxtNotEmpty(records); err != nil {
		return err
	} // as of 2022-07

	if err := recordaudit.TxtNoUnpairedDoubleQuotes(records); err != nil {
		return err
	} // as of 2022-07

	return nil
}
