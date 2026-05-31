package mustbe

import (
	"fmt"
	"strconv"
)

func Bool(a any) bool {
	switch v := a.(type) {
	case bool:
		return v
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			panic(fmt.Sprintf("Bool: invalid boolean string: %s", a))
		}
		return b
	}
	panic(fmt.Sprintf("Bool: unhandled type: %T", a))
}
