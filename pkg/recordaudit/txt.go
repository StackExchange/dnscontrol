package recordaudit

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// Keep these in alphabetical order.

// TxtBackticks audits TXT records for strings that contain backticks.
func TxtBackticks(records []*models.RecordConfig) error {
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

// TxtEmpty audits TXT records for empty strings.
func TxtEmpty(records []*models.RecordConfig) error {
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

// TxtLen255 audits TXT records for strings exactly 255 octets long.
func TxtLen255(records []*models.RecordConfig) error {
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

// TxtTrailingSpace audits TXT records for strings that end with space.
func TxtTrailingSpace(records []*models.RecordConfig) error {
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
