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
