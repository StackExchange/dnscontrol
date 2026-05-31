package privatetypesrdata

import "fmt"

type NETLIFYV6 struct {
}

func (rd NETLIFYV6) Len() int {
	return 0
}

func (rd NETLIFYV6) String() string {
	return ""
}

func MakeNETLIFYV6(orgin string, args ...any) (NETLIFYV6, error) {
	if len(args) != 0 {
		return NETLIFYV6{}, fmt.Errorf("NETLIFYV6: wrong number of arguments, expected 0, got %d: %+v", len(args), args)
	}
	return NETLIFYV6{}, nil
}
