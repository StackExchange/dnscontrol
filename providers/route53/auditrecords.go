package route53

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("R53_ALIAS", rejectifTargetEqualsLabel) // Last verified 2023-03-01
	a.Add("TXT", rejectif.TxtIsEmpty)             // Last verified 2023-10-28

	return a.Audit(records)
}

// Normally this kind of function would be put in `pkg/rejectif` but
// since this is ROUTE53-specific, we'll include it here.

// rejectifTargetEqualsLabel rejects an ALIAS that would create a loop.

func rejectifTargetEqualsLabel(rc *models.RecordConfig) error {
	if (rc.GetLabelFQDN() + ".") == rc.GetTargetField() {
		return fmt.Errorf("alias target loop")
	}
	return nil
}
