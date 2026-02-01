package rtypecontrol

import (
	"fmt"
	"strings"
)

func StringifyQuoted(args []any) string {
	if len(args) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%q", args[0]))
	for _, arg := range args[1:] {
		sb.WriteString(fmt.Sprintf(" %q", arg))
	}
	return sb.String()
}
