package recordaudit

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// Keep these in alphabetical order.

// TxtNoBackticks audits TXT records for strings that contain backticks.
func TxtNoBackticks(rc *models.RecordConfig) error {
	for _, txt := range rc.TxtStrings {
		if strings.Contains(txt, "`") {
			return fmt.Errorf("txtstring contains backtick")
		}
	}
	return nil
}

// TxtNoSingleQuotes audits TXT records for strings that contain single-quotes.
func TxtNoSingleQuotes(rc *models.RecordConfig) error {
	for _, txt := range rc.TxtStrings {
		if strings.Contains(txt, "'") {
			return fmt.Errorf("txtstring contains single-quotes")
		}
	}
	return nil
}

// TxtNoDoubleQuotes audits TXT records for strings that contain doublequotes.
func TxtNoDoubleQuotes(rc *models.RecordConfig) error {
	for _, txt := range rc.TxtStrings {
		if strings.Contains(txt, `"`) {
			return fmt.Errorf("txtstring contains doublequotes")
		}
	}
	return nil
}

// TxtNoStringsExactlyLen255 audits TXT records for strings exactly 255 octets long.
// This is rare; you probably want to use TxtNoLongStrings() instead.
func TxtNoStringsExactlyLen255(rc *models.RecordConfig) error {
	for _, txt := range rc.TxtStrings {
		if len(txt) == 255 {
			return fmt.Errorf("txtstring length is 255")
		}
	}
	return nil
}

// TxtNoStringsLen256orLonger audits TXT records for strings that are >255 octets.
func TxtNoStringsLen256orLonger(rc *models.RecordConfig) error {
	for _, txt := range rc.TxtStrings {
		if len(txt) > 255 {
			return fmt.Errorf("%q txtstring length > 255", rc.GetLabel())
		}
	}
	return nil
}

// TxtNoMultipleStrings audits TXT records for multiple strings
func TxtNoMultipleStrings(rc *models.RecordConfig) error {
	if len(rc.TxtStrings) > 1 {
		return fmt.Errorf("multiple strings in one txt")
	}
	return nil
}

// TxtNoTrailingSpace audits TXT records for strings that end with space.
func TxtNoTrailingSpace(rc *models.RecordConfig) error {
	for _, txt := range rc.TxtStrings {
		if txt != "" && txt[ultimate(txt)] == ' ' {
			return fmt.Errorf("txtstring ends with space")
		}
	}
	return nil
}

// TxtNotEmpty audits TXT records for empty strings.
func TxtNotEmpty(rc *models.RecordConfig) error {
	// There must be strings.
	if len(rc.TxtStrings) == 0 {
		return fmt.Errorf("txt with no strings")
	}
	// Each string must be non-empty.
	for _, txt := range rc.TxtStrings {
		if len(txt) == 0 {
			return fmt.Errorf("txtstring is empty")
		}
	}
	return nil
}

// TxtNoUnpairedDoubleQuotes audits TXT records for strings that contain unpaired doublequotes.
func TxtNoUnpairedDoubleQuotes(rc *models.RecordConfig) error {
	for _, txt := range rc.TxtStrings {
		if strings.Count(txt, `"`)%2 == 1 {
			return fmt.Errorf("txtstring contains unpaired doublequotes")
		}
	}
	return nil
}
