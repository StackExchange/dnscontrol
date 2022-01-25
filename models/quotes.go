package models

import (
	"encoding/csv"
	"strings"
)

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

// ParseQuotedFields is like strings.Fields except individual fields
// might be quoted using `"`.
func ParseQuotedFields(s string) ([]string, error) {
	// Fields are space-separated but a field might be quoted.  This is,
	// essentially, a CSV where spaces are the field separator (not
	// commas). Therefore, we use the CSV parser. See https://stackoverflow.com/a/47489846/71978
	r := csv.NewReader(strings.NewReader(s))
	r.Comma = ' ' // space
	return r.Read()
}
