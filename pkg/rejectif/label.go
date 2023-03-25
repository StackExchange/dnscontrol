package rejectif

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// Keep these in alphabetical order.

// LabelNotApex detects use not at apex. Use this when a record type
// is only permitted at the apex.
func LabelNotApex(rc *models.RecordConfig) error {
	if rc.GetLabel() != "@" {
		return fmt.Errorf("use not at apex")
	}
	return nil
}
