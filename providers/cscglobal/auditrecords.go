package cscglobal

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/recordaudit"
)

// AuditRecords returns an error if any records are not
// supportable by this provider.
func AuditRecords(records []*models.RecordConfig) error {

	// Each test should be encapsulated in a function that can be tested
	// individually.  If the test is of general use, add it to the
	// recordaudit module.

	// Each test should document the last time we verified the test was
	// still needed. Sometimes companies change their API.

	if err := recordaudit.TxtNoDoubleQuotes(records); err != nil {
		return err
	} // Needed as of 2022-08-08

	//	if err := recordaudit.TxtNoStringsLen256orLonger(records); err != nil {
	//		return err
	//	} // Needed as of 2022-06-10

	if err := recordaudit.TxtNoMultipleStrings(records); err != nil {
		return err
	} // Needed as of 2022-06-10

	if err := recordaudit.TxtNoTrailingSpace(records); err != nil {
		return err
	} // Needed as of 2022-06-10

	if err := recordaudit.TxtNotEmpty(records); err != nil {
		return err
	} // Needed as of 2022-06-10

	return nil
}
