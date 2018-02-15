package models

import "strings"

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
// `"foo" "bar"` -> []string{"foo", "bar"}
// NOTE: it is assumed there is exactly one space between the quotes.
func ParseQuotedTxt(s string) []string {
	if !IsQuoted(s) {
		return []string{s}
	}
	return strings.Split(StripQuotes(s), `" "`)
}
