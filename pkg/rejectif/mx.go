package rejectif

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
)

// Keep these in alphabetical order.

// MxNull detects MX records that are a "null MX".
// This is needed by providers that don't support RFC 7505.
func MxNull(rc *models.RecordConfig) error {
	if rc.GetTargetField() == "." {
		return fmt.Errorf("mx has null target")
	}
	return nil
}
