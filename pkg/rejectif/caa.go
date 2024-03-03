package rejectif

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
)

// Keep these in alphabetical order.

// CaaFlagIsNonZero identifies CAA records where tag is no zero.
func CaaFlagIsNonZero(rc *models.RecordConfig) error {
	if rc.CaaFlag != 0 {
		return fmt.Errorf("caa flag is non-zero")
	}
	return nil
}

// CaaTargetContainsWhitespace identifies CAA records that have
// whitespace in the target.
// See https://github.com/StackExchange/dnscontrol/issues/1374
func CaaTargetContainsWhitespace(rc *models.RecordConfig) error {
	if strings.ContainsAny(rc.GetTargetField(), " \t\r\n") {
		return fmt.Errorf("caa target contains whitespace")
	}
	return nil
}

// // CaaTargetHasSemicolon identifies CAA records that contain semicolons.
// func CaaTargetHasSemicolon(rc *models.RecordConfig) error {
// 	if strings.Contains(rc.GetTargetField(), ";") {
// 		return fmt.Errorf("caa target contains semicolon")
// 	}
// 	return nil
// }
