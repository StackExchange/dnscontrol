package rejectif

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// Keep these in alphabetical order.

// CaaTargetHasSemicolon audits CAA records for issues that contain semicolons.
func CaaTargetHasSemicolon(rc *models.RecordConfig) error {
	if strings.Contains(rc.GetTargetField(), ";") {
		return fmt.Errorf("caa target contains semicolon")
	}
	return nil
}
