package rejectif

import (
	"errors"

	"github.com/StackExchange/dnscontrol/v4/models"
)

// Keep these in alphabetical order.

// SrvHasNullTarget detects SRV records that has a null target.
func SrvHasNullTarget(rc *models.RecordConfig) error {
	if rc.GetTargetField() == "." {
		return errors.New("srv has null target")
	}
	return nil
}
