package msdns

import "strings"

func escapePS(s string) string {
	return `'` + strings.Replace(s, `'`, `"`, -1) + `'`
}
