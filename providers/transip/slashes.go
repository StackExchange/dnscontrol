package transip

import (
	"regexp"
)

var removeSlashesRegexp = regexp.MustCompile(`(?:\\(\\)+)|(?:\\)`)

func removeSlashes(s string) string {
	return removeSlashesRegexp.ReplaceAllString(s, "$1")
}
