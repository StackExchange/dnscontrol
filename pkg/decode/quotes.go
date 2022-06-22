package decode

// IsQuoted returns true if the string starts and ends with a double quote.
func IsQuoted(s string) bool {
	if len(s) < 2 {
		return false
	}
	if s[0] == '"' && s[ultimate(s)] == s[0] {
		return true
	}
	return false
}

// StripQuotes returns the string with the starting and ending quotes removed.
// If it is not quoted, the original string is returned.
func StripQuotes(s string) string {
	if IsQuoted(s) {
		return s[1 : len(s)-1]
	}
	return s
}
