package models

import "fmt"

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
	return validateTXT(rc)
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
	return validateTXT(rc)
}

// SetTargetTXTString is like SetTargetTXT but accepts one big string,
// which must be parsed into one or more strings based on how it is quoted.
// Ex: foo             << 1 string
//     foo bar         << 1 string
//     "foo" "bar"     << 2 strings
func (rc *RecordConfig) SetTargetTXTString(s string) error {
	return rc.SetTargetTXTs(ParseQuotedTxt(s))
}

func validateTXT(rc *RecordConfig) error {
	if rc.Type != "TXT" {
		return fmt.Errorf("rc.Type=%q, expecting TXT", rc.Type)
	}
	for i, _ := range rc.TxtStrings {
		l := len(rc.TxtStrings[i])
		if l > 255 {
			return fmt.Errorf("txt TxtString[%d] too long (%d bytes > 255)", i, l)
		}
	}
	l := len(rc.GetTargetField())
	if l > 255 {
		return fmt.Errorf("txt string too long (%d bytes > 255)", l)
	}
	return nil
}
