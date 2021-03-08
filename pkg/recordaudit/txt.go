package recordaudit

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// Keep these in alphabetical order.

// TxtNoBackticks audits TXT records for strings that contain backticks.
func TxtNoBackticks(records []*models.RecordConfig) error {
	for _, rc := range records {

		if rc.HasFormatIdenticalToTXT() {
			for _, txt := range rc.TxtStrings {
				if strings.Index(txt, "`") != -1 {
					return fmt.Errorf("txtstring contains backtick")
				}
			}
		}

	}
	return nil
}

// TxtNoSingleQuotes audits TXT records for strings that contain single-quotes.
func TxtNoSingleQuotes(records []*models.RecordConfig) error {
	for _, rc := range records {

		if rc.HasFormatIdenticalToTXT() {
			for _, txt := range rc.TxtStrings {
				if strings.Index(txt, "'") != -1 {
					return fmt.Errorf("txtstring contains single-quotes")
				}
			}
		}

	}
	return nil
}

// TxtNoDoubleQuotes audits TXT records for strings that contain doublequotes.
func TxtNoDoubleQuotes(records []*models.RecordConfig) error {
	for _, rc := range records {

		if rc.HasFormatIdenticalToTXT() {
			for _, txt := range rc.TxtStrings {
				if strings.Index(txt, `"`) != -1 {
					return fmt.Errorf("txtstring contains doublequotes")
				}
			}
		}

	}
	return nil
}

// TxtNoLen255 audits TXT records for strings exactly 255 octets long.
func TxtNoLen255(records []*models.RecordConfig) error {
	for _, rc := range records {

		if rc.HasFormatIdenticalToTXT() { // TXT and similar:
			for _, txt := range rc.TxtStrings {
				if len(txt) == 255 {
					return fmt.Errorf("txtstring length is 255")
				}
			}
		}

	}
	return nil
}

// TxtNoLongStrings audits TXT records for strings that are >255 octets.
func TxtNoLongStrings(records []*models.RecordConfig) error {
	for _, rc := range records {

		if rc.HasFormatIdenticalToTXT() { // TXT and similar:
			for _, txt := range rc.TxtStrings {
				if len(txt) > 255 {
					return fmt.Errorf("txtstring length > 255")
				}
			}
		}

	}

	return nil
}

// TxtNoMultipleStrings audits TXT records for multiple strings
func TxtNoMultipleStrings(records []*models.RecordConfig) error {
	for _, rc := range records {

		if rc.HasFormatIdenticalToTXT() { // TXT and similar:
			if len(rc.TxtStrings) > 1 {
				return fmt.Errorf("multiple strings in one txt")
			} else if len(rc.TxtStrings) == 1 && len(rc.TxtStrings[0]) > 255 {
				return fmt.Errorf("strings >255 octets")
			}
		}

	}
	return nil
}

// TxtNoTrailingSpace audits TXT records for strings that end with space.
func TxtNoTrailingSpace(records []*models.RecordConfig) error {
	for _, rc := range records {

		if rc.HasFormatIdenticalToTXT() { // TXT and similar:
			for _, txt := range rc.TxtStrings {
				if txt != "" && txt[ultimate(txt)] == ' ' {
					return fmt.Errorf("txtstring ends with space")
				}
			}
		}

	}
	return nil
}

// TxtNotEmpty audits TXT records for empty strings.
func TxtNotEmpty(records []*models.RecordConfig) error {
	for _, rc := range records {

		if rc.HasFormatIdenticalToTXT() { // TXT and similar:
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
		}

	}
	return nil
}
