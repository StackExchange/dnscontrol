package mustbe

import (
	"fmt"
)

func RawString(a any) string {
	switch v := a.(type) {
	case string:
		return v
	}
	return fmt.Sprintf("%s", a)

}
