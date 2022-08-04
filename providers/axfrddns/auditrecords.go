package axfrddns

import "github.com/StackExchange/dnscontrol/v3/models"

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	return nil
}
