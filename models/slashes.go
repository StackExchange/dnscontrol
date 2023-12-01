package models

import (
	"regexp"
)

func RemoveSlashes(s string) string {
	m := regexp.MustCompile("(?:\\\\(\\\\)+)|(?:\\\\)")
	return m.ReplaceAllString(s, "$1")
}
