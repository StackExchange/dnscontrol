package models

import "strings"

/*
// SetTxt sets the value of a TXT record to s.
func (rc *RecordConfig) SetTxt(s string) {
	rc.Target = s
	rc.TxtStrings = []string{s}
}
*/

/*
// SetTxts sets the value of a TXT record to the list of strings s.
func (rc *RecordConfig) SetTxts(s []string) {
	rc.Target = s[0]
	rc.TxtStrings = s
}
*/

// SetTxtParse sets the value of TXT record if the list of strings is combined into one string.
// `foo`  -> []string{"foo"}
// `"foo"` -> []string{"foo"}
// `"foo" "bar"` -> []string{"foo" "bar"}
func (rc *RecordConfig) SetTxtParse(s string) {
	ss := ParseQuotedTxt(s)
	rc.Target = ss[0]
	rc.TxtStrings = ss
}

// IsQuoted returns true if the string starts and ends with a double quote.
func IsQuoted(s string) bool {
	if s == "" {
		return false
	}
	if len(s) < 2 {
		return false
	}
	if s[0] == '"' && s[len(s)-1] == s[0] {
		return true
	}
	return false
}

// StripQuotes returns the string with the starting and ending quotes removed.
func StripQuotes(s string) string {
	if IsQuoted(s) {
		return s[1 : len(s)-1]
	}
	return s
}

// ParseQuotedTxt returns the individual strings of a combined quoted string.
// `foo`  -> []string{"foo"}
// `"foo"` -> []string{"foo"}
// `"foo" "bar"` -> []string{"foo" "bar"}
// NOTE: it is assumed there is exactly one space between the quotes.
func ParseQuotedTxt(s string) []string {
	if !IsQuoted(s) {
		return []string{s}
	}
	return strings.Split(StripQuotes(s), `" "`)
}
