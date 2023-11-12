//go:generate stringer -type=State

package txtutil

import (
	"bytes"
	"fmt"
	"strings"
)

func ParseQuoted(s string) (string, error) {
	return txtDecode(s)
}

func EncodeQuoted(t string) string {
	return txtEncode(ToChunks(t))
}

type State int

const (
	StateStart     State = iota // Looking for a non-space
	StateUnquoted               // A run of unquoted text
	StateQuoted                 // Quoted text
	StateBackslash              // last char was backlash in a quoted string
	StateWantSpace              // expect space after closing quote
)

func isRemaining(s string, i, r int) bool {
	return (len(s) - 1 - i) > r
}

// txtDecode decodes TXT strings received from ROUTE53 and GCLOUD.
func txtDecode(s string) (string, error) {
	// Parse according to RFC1035 zonefile specifications.
	// "foo"  -> one string: `foo``
	// "foo" "bar"  -> two strings: `foo` and `bar`
	// quotes and backslashes are escaped using \

	//printer.Printf("DEBUG: route53 txt inboundv=%v\n", s)

	b := &bytes.Buffer{}
	state := StateStart
	for i, c := range s {

		//printer.Printf("DEBUG: state=%v rune=%v\n", state, string(c))

		switch state {

		case StateStart:
			if c == ' ' {
				// skip whitespace
			} else if c == '"' {
				state = StateQuoted
			} else {
				state = StateUnquoted
				b.WriteRune(c)
			}

		case StateUnquoted:

			if c == ' ' {
				state = StateStart
			} else {
				b.WriteRune(c)
			}

		case StateQuoted:

			if c == '\\' {
				if isRemaining(s, i, 1) {
					state = StateBackslash
				} else {
					return "", fmt.Errorf("txtDecode quoted string ends with backslash q(%q)", s)
				}
			} else if c == '"' {
				state = StateWantSpace
			} else {
				b.WriteRune(c)
			}

		case StateBackslash:
			b.WriteRune(c)
			state = StateQuoted

		case StateWantSpace:
			if c == ' ' {
				state = StateStart
			} else {
				return "", fmt.Errorf("txtDecode expected whitespace after close quote q(%q)", s)
			}

		}
	}

	r := b.String()
	//printer.Printf("DEBUG: route53 txt decodedv=%v\n", r)
	return r, nil
}

// txtEncode encodes TXT strings as expected by ROUTE53 and GCLOUD.
func txtEncode(ts []string) string {
	//printer.Printf("DEBUG: route53 txt outboundv=%v\n", ts)

	for i := range ts {
		ts[i] = strings.ReplaceAll(ts[i], `\`, `\\`)
		ts[i] = strings.ReplaceAll(ts[i], `"`, `\"`)
	}
	t := `"` + strings.Join(ts, `" "`) + `"`

	//printer.Printf("DEBUG: route53 txt  encodedv=%v\n", t)
	return t
}
