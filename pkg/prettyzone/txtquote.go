package prettyzone

import (
	"strconv"
	"strings"
)

func txtToNative(parts []string) string {
	var quotedParts []string

	for _, part := range parts {
		quotedParts = append(quotedParts, strconv.Quote(part))
	}

	return strings.Join(quotedParts, " ")
}
