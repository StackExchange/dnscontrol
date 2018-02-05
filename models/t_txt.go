package models

// SetTargetTXT sets the TXT fields when there is 1 string.
func (rc *RecordConfig) SetTargetTXT(s string) error {
	rc.Target = s
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
	rc.Target = s[0]
	rc.TxtStrings = s
	if rc.Type == "" {
		rc.Type = "TXT"
	}
	if rc.Type != "TXT" {
		panic("assertion failed: SetTargetTXT called when .Type is not TXT")
	}
	return nil
}

// SetTargetTXTString is like SetTargetTXT but accepts one big string.
// Ex: foo             << 1 string
//     foo bar         << 1 string
//     "foo" "bar"     << 2 strings
func (rc *RecordConfig) SetTargetTXTString(s string) error {
	return rc.SetTargetTXTs(ParseQuotedTxt(s))
}
