package rfc4183

import (
	"fmt"
	"strings"
)

var newmode bool
var modeset bool

func SetCompatibilityMode(m string) error {
	if modeset {
		return fmt.Errorf("ERROR: REVCOMPAT() already set")
	}
	modeset = true

	switch strings.ToLower(m) {
	case "rfc2317", "2317", "2", "old":
		newmode = false
	case "rfc4183", "4183", "4":
		newmode = true
	default:
		return fmt.Errorf("invalid value %q, must be rfc2317 or rfc4182", m)
	}
	return nil
}

func IsRFC4183Mode() bool {
	return newmode
}

var warningNeeded bool = false

func NeedsWarning() {
	warningNeeded = true
}

func PrintWarning() {
	if modeset {
		// No warnings if REVCOMPAT() was used.
		return
	}
	if !warningNeeded {
		return
	}
	fmt.Printf("WARNING: REV() breaking change coming in v5.0. See https://docs.dnscontrol.org/functions/REVCOMPAT\n")
}
