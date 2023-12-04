package transip

import (
	"regexp"
)

func removeSlashes(s string) string {
	m := regexp.MustCompile("(?:\\\\(\\\\)+)|(?:\\\\)")
	return m.ReplaceAllString(s, "$1")
}
