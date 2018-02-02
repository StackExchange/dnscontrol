package models

// SetTargetTXT sets the TXT fields when there is 1 string.
func (rc *RecordConfig) SetTargetTXT(s string) {
	rc.SetTxt(s)
	if rc.Type == "" {
		rc.Type = "TXT"
	}
	if rc.Type != "TXT" {
		panic("SetTargetTXT called when .Type is not TXT")
	}
}

// SetTargetTXTs sets the TXT fields when there are many strings.
func (rc *RecordConfig) SetTargetTXTs(s []string) {
	rc.SetTxts(s)
	if rc.Type == "" {
		rc.Type = "TXT"
	}
	if rc.Type != "TXT" {
		panic("SetTargetTXT called when .Type is not TXT")
	}
}

// SetTargetTXTString is like SetTargetTXT but accepts one big string.
func (rc *RecordConfig) SetTargetTXTString(s string) {
	rc.SetTxtParse(s)
}

// Helper functions:

// SetTxt sets the value of a TXT record to s.
func (rc *RecordConfig) SetTxt(s string) {
	rc.Target = s
	rc.TxtStrings = []string{s}
}

// SetTxts sets the value of a TXT record to the list of strings s.
func (rc *RecordConfig) SetTxts(s []string) {
	rc.Target = s[0]
	rc.TxtStrings = s
}

// SetTxtParse sets the value of TXT record if the list of strings is combined into one string.
// `foo`  -> []string{"foo"}
// `"foo"` -> []string{"foo"}
// `"foo" "bar"` -> []string{"foo" "bar"}
func (rc *RecordConfig) SetTxtParse(s string) {
	rc.SetTxts(ParseQuotedTxt(s))
}
