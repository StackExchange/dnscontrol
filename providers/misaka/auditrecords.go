package misaka

import (
	"fmt"

	"github.com/DNSControl/dnscontrol/v5/models"
	"github.com/DNSControl/dnscontrol/v5/pkg/rejectif"
)

// maxTTL is the largest TTL the API accepts.
const maxTTL = 2147483647

// AuditRecords returns a list of errors corresponding to the records
// that aren't supported by this provider.  If all records are
// supported, an empty list is returned.
func (*misakaProvider) AuditRecords(records models.Records) []error {
	a := rejectif.Auditor{}

	// 422 Ensure this value is less than or equal to 2147483647.
	// Last verified 2026-10-14.
	a.Add("*", rejectTTLAboveMax)

	return a.Audit(records)
}

// rejectTTLAboveMax rejects a TTL the API would refuse, instead of quietly
// lowering it to one the user did not ask for.
func rejectTTLAboveMax(rc *models.RecordConfig) error {
	if rc.TTL > maxTTL {
		return fmt.Errorf("%s %s: ttl %d is larger than %d, the maximum Misaka.io accepts", rc.GetLabelFQDN(), rc.Type, rc.TTL, maxTTL)
	}
	return nil
}
