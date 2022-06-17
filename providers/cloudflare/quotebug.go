package cloudflare

import "strings"

func isCloudflareQuoteBug(s string) bool {
	if len(s) < 5 {
		return false
	}
	if s[0] != '"' {
		return false
	}
	return !strings.Contains(s[1:len(s)-1], `" "`)
}

func fixCloudflareQuoteBug(s string) string {
	s = strings.Replace(s, `\`, `\\`, -1)
	s = strings.Replace(s, `"`, `\"`, -1)
	return `"` + s + `"`
}
