package cfsingleredirect

import (
	"fmt"
)

func makeSingleRedirectTarget(name string, code uint16, when, then string) string {
	return fmt.Sprintf("%s code=(%03d) when=(%s) then=(%s)",
		name,
		code,
		when,
		then,
	)
}

func mkTargetAPI(name string, code uint16, when, then string) string {
	return fmt.Sprintf("%s code=(%03d) when=(%s) then=(%s)",
		//return fmt.Sprintf("%s when=(%s) then=(%s)",
		name,
		code,
		when,
		then,
	)
}
