package mustbe

import (
	"fmt"
)

func Host(origin string, a any) string {
	switch v := a.(type) {
	case string:
		return v // FIXME: this should be cleaned up to be a proper hostname, but for now we just want to get the tests working.
	}
	panic(fmt.Sprintf("Host: unhandled type: %T", a))
}
