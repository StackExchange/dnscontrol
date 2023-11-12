//go:generate stringer -type=State

package route53

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
)

type State int

const (
	StateStart            State = iota // Normal text
	StateQuoted                        // Quoted text
	StateBackslash                     // last char was backslash
	StateQuotedBackslash               // last char was backlash in a quoted string
	StateWantSpace                     // expect space after closing quote
	StateWantSpaceOrQuote              // expect open quote after `" `
)

func isRemaining(s string, i, r int) bool {
	return (len(s) - 1 - i) > r
}

// txtDecode is like strings.Fields except individual fields
// might be quoted using `"`.
func txtDecode(s string) (string, error) {
	printer.Printf("DEBUG: route53 txt inboundv=%v\n", s)
	// Parse according to RFC1035 zonefile specifications.
	// "foo"  -> one string: `foo``
	// "foo" "bar"  -> two strings: `foo` and `bar`

	// if s == `""` {
	// 	r := []string{}
	// 	printer.Printf("DEBUG: route53 txt Z decodedv=%v\n", r)
	// 	return r, nil
	// }

	b := &bytes.Buffer{}
	state := StateStart
	for i, c := range s {

		//printer.Printf("DEBUG: state=%v rune=%v\n", state, string(c))

		switch state {

		case StateStart:

			if c == '"' {
				state = StateQuoted
			} else if c == ' ' {
				state = StateQuoted
			} else if c == '\\' {
				if isRemaining(s, i, 1) {
					state = StateBackslash
				} else {
					return "", fmt.Errorf("txtDecode string ends with backslash q(%q)", s)
				}
			} else {
				b.WriteRune(c)
			}

		case StateBackslash:
			b.WriteRune(c)
			state = StateStart

		case StateQuoted:

			if c == '\\' {
				if isRemaining(s, i, 1) {
					state = StateQuotedBackslash
				} else {
					return "", fmt.Errorf("txtDecode quoted string ends with backslash q(%q)", s)
				}
			} else if c == '"' {
				state = StateWantSpace
			} else {
				b.WriteRune(c)
			}

		case StateQuotedBackslash:
			b.WriteRune(c)
			state = StateQuoted

		case StateWantSpace:
			if c == ' ' {
				state = StateWantSpaceOrQuote
			} else {
				return "", fmt.Errorf("txtDecode expected whitespace after close quote q(%q)", s)
			}

		case StateWantSpaceOrQuote:
			if c == ' ' {
				state = StateWantSpaceOrQuote
			} else if c == '"' {
				state = StateQuoted
			} else {
				state = StateStart
				b.WriteRune(c)
			}

		}
	}

	r := b.String()
	printer.Printf("DEBUG: route53 txt decodedv=%v\n", r)
	return r, nil
}

func txtEncode(ts []string) string {
	printer.Printf("DEBUG: route53 txt outboundv=%v\n", ts)

	if len(ts) == 0 {
		t := `""`
		printer.Printf("DEBUG: route53 txt Z  encodedv=%v\n", t)
		return t
	}

	for i := range ts {
		ts[i] = strings.ReplaceAll(ts[i], `\`, `\\`)
		ts[i] = strings.ReplaceAll(ts[i], `"`, `\"`)
	}
	t := `"` + strings.Join(ts, `" "`) + `"`

	printer.Printf("DEBUG: route53 txt  encodedv=%v\n", t)
	return t
}
