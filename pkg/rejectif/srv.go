package rejectif

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
)

// Keep these in alphabetical order.

// SrvHasNullTarget detects SRV records that has a null target.
func SrvHasNullTarget(rc *models.RecordConfig) error {
	if rc.GetTargetField() == "." {
		return fmt.Errorf("srv has null target")
	}
	return nil
}
