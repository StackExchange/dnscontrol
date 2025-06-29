package adguardhome

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

	a.Add("ALIAS", rejectif.LabelNotApex)

	a.Add("ADGUARDHOME_A_PASSTHROUGH", nonNullValue)

	a.Add("ADGUARDHOME_AAAA_PASSTHROUGH", nonNullValue)

	return a.Audit(records)
}

func nonNullValue(v *models.RecordConfig) error {
	if len(v.GetTargetField()) != 0 {
		return fmt.Errorf("%s rtype value should be empty", v.Type)
	}

	return nil
}
