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

// TxtHasSegmentLen256orLonger audits TXT records for strings that are >255 octets.
func TxtHasSegmentLen256orLonger(rc *models.RecordConfig) error {
	if len(rc.GetTargetTXTJoined()) > 255 {
		return fmt.Errorf("%q txtstring length > 255", rc.GetLabel())
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

// TxtHasMultipleSegments audits TXT records for multiple strings
func TxtHasMultipleSegments(rc *models.RecordConfig) error {
	if len(rc.GetTargetTXTSegmented()) > 1 {
		return fmt.Errorf("multiple strings in one txt")
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

// TxtIsExactlyLen255 audits TXT records for strings exactly 255 octets long.
// This is rare; you probably want to use TxtNoStringsLen256orLonger() instead.
func TxtIsExactlyLen255(rc *models.RecordConfig) error {
	if len(rc.GetTargetTXTJoined()) == 255 {
		return fmt.Errorf("txtstring length is 255")
	}
	return nil
}

// TxtLongerThan255 audits TXT records for multiple strings
func TxtLongerThan255(rc *models.RecordConfig) error {
	if len(rc.GetTargetTXTJoined()) > 255 {
		return fmt.Errorf("multiple strings in one txt")
	}
	return nil
}
