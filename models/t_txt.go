package models

import (
	"fmt"
	"strings"
)

// SetTargetTXT sets the TXT fields when there is 1 string.
func (rc *RecordConfig) SetTargetTXT(s string) error {
	rc.SetTarget(s)
	rc.TxtStrings = []string{s}
	if rc.Type == "" {
		rc.Type = "TXT"
	}
	if rc.Type != "TXT" {
		panic("assertion failed: SetTargetTXT called when .Type is not TXT")
	}
	return nil
}

// SetTargetTXTs sets the TXT fields when there are many strings.
func (rc *RecordConfig) SetTargetTXTs(s []string) error {
	rc.SetTarget(s[0])
	rc.TxtStrings = s
	if rc.Type == "" {
		rc.Type = "TXT"
	}
	if rc.Type != "TXT" {
		panic("assertion failed: SetTargetTXT called when .Type is not TXT")
	}
	return nil
}

// SetTargetTXTString is like SetTargetTXT but accepts one big string,
// which must be parsed into one or more strings based on how it is quoted.
// Ex: foo             << 1 string
//     foo bar         << 1 string
//     "foo" "bar"     << 2 strings
func (rc *RecordConfig) SetTargetTXTString(s string) error {
	return rc.SetTargetTXTs(ParseQuotedTxt(s))
}

// TxtNormalize splits long txt targets if required based on the algo.
func (rc *RecordConfig) TxtNormalize(algo string) {
	switch algo {
	case "multistring":
		rc.SetTargetTXTs(splitChunks(strings.Join(rc.TxtStrings, ""), 255))
	case "space":
		panic("not implemented")
	default:
		panic("TxtNormalize called with invalid algorithm")
	}
}

func splitChunks(buf string, lim int) []string {
	var chunk string
	chunks := make([]string, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}

// ValidateTXT returns an error if the txt record is invalid.
// Verifies the Target and TxtStrings are less than 255 bytes each.
func ValidateTXT(rc *RecordConfig) error {
	if rc.Type != "TXT" {
		return fmt.Errorf("rc.Type=%q, expecting TXT", rc.Type)
	}
	for i := range rc.TxtStrings {
		l := len(rc.TxtStrings[i])
		if l > 255 {
			return fmt.Errorf("txt target >255 bytes and AUTOSPLIT not set: label=%q index=%d len=%d string[:50]=%q", rc.GetLabel(), i, l, rc.TxtStrings[i][:50]+"...")
		}
	}
	return nil
}
