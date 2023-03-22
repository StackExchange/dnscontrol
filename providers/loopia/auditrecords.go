package loopia

import (
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/rejectif"
)

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func AuditRecords(records []*models.RecordConfig) []error {
	a := rejectif.Auditor{}

	a.Add("TXT", rejectif.TxtIsEmpty) // Last verified 2023-03-10: Loopia returns 404

	//Loopias TXT length limit appears to be 450 octets
	a.Add("TXT", TxtHasSegmentLen450orLonger)

	a.Add("MX", rejectif.MxNull) // Last verified 2023-03-23

	return a.Audit(records)
}

// TxtHasSegmentLen450orLonger audits TXT records for strings that are >450 octets.
func TxtHasSegmentLen450orLonger(rc *models.RecordConfig) error {
	for _, txt := range rc.TxtStrings {
		if len(txt) > 450 {
			return fmt.Errorf("%q txtstring length > 450", rc.GetLabel())
		}
	}
	return nil
}
