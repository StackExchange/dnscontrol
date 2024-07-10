package rtypecontrol

import (
	"fmt"
	"strconv"
)

// PaveArgs converts each arg to its desired type, or returns an error if conversion fails or if the number of arguments is wrong.
// argTypes is a string where each rune specifies the desired type of the arg in the same position:
// 'i': uinet16 (will convert strings, truncate floats, etc)
// 's': Valid only if string.
func PaveArgs(args []any, argTypes string) error {

	if len(args) != len(argTypes) {
		return fmt.Errorf("wrong number of arguments. Expected %v, got %v", len(argTypes), len(args))
	}

	for i, at := range argTypes {
		arg := args[i]
		switch at {

		case 'i': // uint16
			if s, ok := arg.(string); ok { // Is this a string-encoded int?
				ni, err := strconv.Atoi(s)
				if err != nil {
					return fmt.Errorf("value %q is not a number (uint16 wanted)", arg)
				}
				args[i] = uint16(ni)
			} else if _, ok := arg.(float64); ok {
				args[i] = uint16(arg.(float64))
			} else if _, ok := arg.(uint16); ok {
				args[i] = arg.(uint16)
			} else if _, ok := arg.(int); ok {
				args[i] = uint16(arg.(int))
			} else {
				return fmt.Errorf("value %q is type %T, expected uint16", arg, arg)
			}

		case 's':
			if _, ok := arg.(string); ok {
				args[i] = arg.(string)
			} else {
				args[i] = fmt.Sprintf("%v", arg)
			}

		}
	}

	return nil
}
