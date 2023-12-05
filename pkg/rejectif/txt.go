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

// TxtLongerThan255 audits TXT records for multiple strings
func TxtLongerThan255(rc *models.RecordConfig) error {
	if len(rc.GetTargetTXTJoined()) > 255 {
		return fmt.Errorf("TXT records longer than 255 octets (chars)")
	}
	return nil
}

func TxtLongerThan(rc *models.RecordConfig, l int) func(rc *models.RecordConfig) error {

	return func(rc *models.RecordConfig) error {
		le := l
		if len(rc.GetTargetTXTJoined()) > l)
		return fmt.Errorf("TXT records longer than xx octets (chars)")
	}
	}
}