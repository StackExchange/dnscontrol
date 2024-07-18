package cfsingleredirect

import (
	"fmt"
)

func MakePageRuleBlob(from, to string, priority, code uint16) string {
	return fmt.Sprintf("%d,%03d,%s,%s", // $PRIO,$CODE,$FROM,$TO
		priority,
		code,
		from,
		to,
	)
}

func MakeSingleRedirectTarget(name string, code uint16, when, then string) string {
	return fmt.Sprintf("name=(%s) code=(%03d) when=(%s) then=(%s)",
		name,
		code,
		when,
		then,
	)
}
