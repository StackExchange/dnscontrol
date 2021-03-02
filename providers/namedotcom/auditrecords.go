package namedotcom

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/recordaudit"
)

// AuditRecords returns an error if any records are not
// supportable by this provider.
func AuditRecords(records []*models.RecordConfig) error {

	if err := MaxLengthNDC(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-01

	if err := recordaudit.TxtNotEmpty(records); err != nil {
		return err
	}
	// Still needed as of 2021-03-01

	return nil
}

// MaxLengthNDC returns and error if the sum of the strings
// are longer than permitted by DigitalOcean. Sadly their
// length limit is undocumented. This seems to work.
func MaxLengthNDC(records []*models.RecordConfig) error {
	for _, rc := range records {

		if rc.HasFormatIdenticalToTXT() { // TXT and similar:
			if len(rc.GetTargetField()) > 512 {
				return fmt.Errorf("encoded txt too long")
			}
		}

	}
	return nil
}
