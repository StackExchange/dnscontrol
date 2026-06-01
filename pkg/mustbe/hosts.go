package mustbe

import (
	"fmt"
	"strings"

	"github.com/DNSControl/dnscontrol/v4/pkg/domaintags"
)

// TargetHost returns a FQDN (or @) suitable as a target for CNAME and other
// records.  origin must be a FQDN without a trailing dot.   arg may be a string or it will be converted to a string.
// * Unicode is converted to PunyCode.
// * The result always ends with a "." unless it is "@".
// * It does not try to turn a FQDN into a shortname, but it will replace the origin with "@". The reason for not shortening it is that "preview" output is unclear when the user sees a shortname. Explicit is better than implicit.
// * This does not handle "*" (wildcards) since they are not valid in targets. That's why this is called TargetHost and not Host.
// Examples: (assume $origin = "domain.com")
// * `@` -> `@`
// * `$origin` -> `@`
// * `foo.$origin.` -> `foo.$origin.`
// * `short` -> `short.$origin`
// * `other.com.` -> `other.com.`
func TargetHost(origin string, arg any) string {
	if strings.HasSuffix(origin, ".") {
		panic("mustbe.Host called with origin ending with .")
	}

	var name string
	switch v := arg.(type) {
	case string:
		name = v
	case int:
		name = fmt.Sprintf("%d", arg)
	default:
		name = fmt.Sprintf("%v", arg)
	}

	// Special symbols:
	switch name {
	case "@":
		return name
	case "":
		return "@"
	}

	// Normalize it
	name = domaintags.EfficientToASCII(name)

	// shorten origin to "@".
	if name == origin+"." {
		return "@"
	}

	// Add domain if needed.
	if strings.HasSuffix(name, ".") {
		return name
	}

	return name + "." + origin + "."
}
