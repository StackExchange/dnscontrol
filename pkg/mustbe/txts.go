package mustbe

import (
	"bytes"
	"fmt"
)

func Txts(args ...any) []string {
	if len(args) == 0 {
		return []string{}
	}
	if len(args) == 1 {
		return []string{fmt.Sprintf("%s", args[0])}
	}
	// Use a string builder. For each args if it's a string, add it to the builder. If it is anything else, convert it to a string and add it to the builder.
	var sb bytes.Buffer
	for _, a := range args {
		sb.WriteString(fmt.Sprintf("%s", a))
	}
	return []string{sb.String()}
}
