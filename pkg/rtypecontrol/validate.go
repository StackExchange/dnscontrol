package rtypecontrol

import (
	"fmt"
	"strconv"
)

// CheckArgTypes validates that the items in args are of appropriate types. argTypes is a string: "ssi" means the args should be string, string, int.
// 's': Valid only if string.
// 'i': Valid only if int, float64, or a string that Atoi() can convert to an int.
func CheckArgTypes(args []any, argTypes string) error {

	if len(args) != len(argTypes) {
		return fmt.Errorf("wrong number of arguments. Expected %v, got %v", len(argTypes), len(args))
	}

	for i, at := range argTypes {
		arg := args[i]
		switch at {

		case 'i':
			if s, ok := arg.(string); ok { // Is this a string-encoded int?
				ni, err := strconv.Atoi(s)
				if err != nil {
					return fmt.Errorf("value %q is type %T, expected INT", arg, arg)
				}
				args[i] = ni
			} else if _, ok := arg.(float64); ok {
				args[i] = int(arg.(float64))
			} else if _, ok := arg.(int); !ok {
				return fmt.Errorf("value %q is type %T, expected INT", arg, arg)
			}

		case 's':
			if _, ok := arg.(string); !ok {
				return fmt.Errorf("value %q is type %T, expected STRING", arg, arg)
			}

		}
	}

	return nil
}
