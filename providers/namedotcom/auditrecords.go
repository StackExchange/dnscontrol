package namedotcom

import (
	"fmt"
	"strings"

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
// are longer than permitted by NDC. Sadly their
// length limit is undocumented. This seems to work.
func MaxLengthNDC(records []*models.RecordConfig) error {
	for _, rc := range records {

		if rc.HasFormatIdenticalToTXT() { // TXT and similar:
			// Sum the length of the segments:
			sum := 0
			for _, segment := range rc.TxtStrings {
				sum += len(segment)                // The length of each segment
				sum += strings.Count(segment, `"`) // Add 1 for any char to be escaped
			}
			// Add the overhead of quoting them:
			n := len(rc.TxtStrings)
			if n > 0 {
				sum += 2 + 3*(n-1) // Start and end double-quotes, plus `" "` between each segment.
			}
			if sum > 512 {
				return fmt.Errorf("encoded txt too long")
			}
		}

	}
	return nil
}
