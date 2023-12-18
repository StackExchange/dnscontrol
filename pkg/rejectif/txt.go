package rejectif

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
)

// Keep these in alphabetical order.

// TxtHasBackslash audits TXT records for strings that contains one or more backslashes.
func TxtHasBackslash(rc *models.RecordConfig) error {
	if strings.Contains(rc.GetTargetTXTJoined(), `\`) {
		return fmt.Errorf("txtstring contains backslashes")
	}
	return nil
}

// TxtHasBackticks audits TXT records for strings that contain backticks.
func TxtHasBackticks(rc *models.RecordConfig) error {
	if strings.Contains(rc.GetTargetTXTJoined(), "`") {
		return fmt.Errorf("txtstring contains backtick")
	}
	return nil
}

// TxtHasDoubleQuotes audits TXT records for strings that contain doublequotes.
func TxtHasDoubleQuotes(rc *models.RecordConfig) error {
	if strings.Contains(rc.GetTargetTXTJoined(), `"`) {
		return fmt.Errorf("txtstring contains doublequotes")
	}
	return nil
}

// TxtHasSemicolon audits TXT records for strings that contain backticks.
func TxtHasSemicolon(rc *models.RecordConfig) error {
	if strings.Contains(rc.GetTargetTXTJoined(), ";") {
		return fmt.Errorf("txtstring contains semicolon")
	}
	return nil
}

// TxtHasSingleQuotes audits TXT records for strings that contain single-quotes.
func TxtHasSingleQuotes(rc *models.RecordConfig) error {
	if strings.Contains(rc.GetTargetTXTJoined(), "'") {
		return fmt.Errorf("txtstring contains single-quotes")
	}
	return nil
}

// TxtHasTrailingSpace audits TXT records for strings that end with space.
func TxtHasTrailingSpace(rc *models.RecordConfig) error {
	txt := rc.GetTargetTXTJoined()
	if txt != "" && txt[ultimate(txt)] == ' ' {
		return fmt.Errorf("txtstring ends with space")
	}
	return nil
}

// TxtHasUnpairedDoubleQuotes audits TXT records for strings that contain unpaired doublequotes.
func TxtHasUnpairedDoubleQuotes(rc *models.RecordConfig) error {
	if strings.Count(rc.GetTargetTXTJoined(), `"`)%2 == 1 {
		return fmt.Errorf("txtstring contains unpaired doublequotes")
	}
	return nil
}

// TxtIsEmpty audits TXT records for empty strings.
func TxtIsEmpty(rc *models.RecordConfig) error {
	if len(rc.GetTargetTXTJoined()) == 0 {
		return fmt.Errorf("txtstring is empty")
	}
	return nil
}

// TxtLongerThan returns a function that audits TXT records for length
// greater than maxLength.
func TxtLongerThan(maxLength int) func(rc *models.RecordConfig) error {
	return func(rc *models.RecordConfig) error {
		m := maxLength
		if len(rc.GetTargetTXTJoined()) > m {
			return fmt.Errorf("TXT records longer than %d octets (chars)", m)
		}
		return nil
	}
}

// TxtStartsOrEndsWithSpaces audits TXT records that starts or ends with spaces
func TxtStartsOrEndsWithSpaces(rc *models.RecordConfig) error {
	txt := rc.GetTargetTXTJoined()
	if len(txt) > 0 && (txt[0] == ' ' || txt[len(txt)-1] == ' ') {
		return fmt.Errorf("txtstring starts or ends with spaces")
	}
	return nil
}
