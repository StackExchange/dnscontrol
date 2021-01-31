package bind

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// makeFileName returns the zone's filename.
func makeFileName(format, uniquename, domain, tag string) string {
	if format == "" {
		fmt.Fprintf(os.Stderr, "BUG: makeFileName called with null format\n")
		return uniquename
	}

	var b bytes.Buffer

	tokens := strings.Split(format, "")
	lastpos := len(tokens) - 1
	for pos := 0; pos < len(tokens); pos++ {
		tok := tokens[pos]
		if tok != "%" {
			b.WriteString(tok)
			continue
		}
		if pos == lastpos {
			b.WriteString("%(string may not end in %)")
			continue
		}
		pos++
		tok = tokens[pos]
		switch tok {
		case "D":
			b.WriteString(domain)
		case "T":
			b.WriteString(tag)
		case "U":
			b.WriteString(uniquename)
		case "?":
			if pos == lastpos {
				b.WriteString("%(string may not end in %?)")
				continue
			}
			pos++
			tok = tokens[pos]
			if tag != "" {
				b.WriteString(tok)
			}
		default:
			fmt.Fprintf(&b, "%%(unknown %%verb %%%s)", tok)
		}
	}

	return b.String()
	//	return strings.Replace(strings.ToLower(uniquename), "/", "_", -1) + ".zone"
}

//// makeFileRegex returns a regex that extracts the domain name from a filename.
//func makeFileRegex(formatstring ) *Regexp {
//	return MustCompile(`(.*).zone`)
//}
