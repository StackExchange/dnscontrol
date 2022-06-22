package decode

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/miekg/dns"
)

// QuotedFields decodes many strings encoded as quoted strings,
// separated by a single space, with no escaping. APIs that encode
// strings this way will have problems with strings that have internal
// quotes.
// Input:               Output:
// `foo`                error
// `one two`            error
// `one "two"`          error
// `"foo"`              []string{`foo`}
// `"foo" "bar"`        []string{`foo`, `bar`}
// `"es\"caped"`        []string{`es\"caped`}
// `"bumble\bee"`       []string{`bumble\bee`}
// `""doublequoted""`   []string{`"doublequoted"`}
// `"\"escquoted\""`    []string{`\"escquoted\"`}
// `"in"side"`          undefined
// `"do""ble"`          undefined
func QuotedFields(s string) ([]string, error) {
	if !IsQuoted(s) {
		// If you get this error, you might use be calling
		// SetTargetTXTfromQuotedFields() when you should use
		// SetTargetTXT().
		return nil, fmt.Errorf("encoding error: no quotes surrounding (%q)", s)
	}
	return strings.Split(StripQuotes(s), `" "`), nil
}

// finds a string surrounded by quotes that might contain an escaped quote character.
var quotedStringRegexp = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)

// QuoteEscapedFields decodes many strings encoded as quoted
// strings, separated by a single space, with internal quotes escaped.
// Input:               Output:
// `foo`                error
// `one two`            error
// `one "two"`          error
// `"foo"`              []string{`foo`}
// `"foo" "bar"`        []string{`foo`, `bar`}
// `"es\"caped"`        []string{`es"caped`}
// `"bumble\bee"`       []string{`bumble\bee`}
// `""doublequoted""`   undefined
// `"\"escquoted\""`    []string{`"escquoted"`}
// `"in"side"`          undefined
// `"do""ble"`          undefined
func QuoteEscapedFields(s string) ([]string, error) {
	if !IsQuoted(s) {
		// If you get this error, you might use be calling
		// SetTargetTXTfromQuotedFields() when you should use
		// SetTargetTXT().
		return nil, fmt.Errorf("encoding error: no quotes surrounding (%q)", s)
	}

	txtStrings := []string{}
	for _, t := range quotedStringRegexp.FindAllStringSubmatch(s, -1) {
		txtString := strings.Replace(t[1], `\"`, `"`, -1)
		txtStrings = append(txtStrings, txtString)
	}
	return txtStrings, nil

}

// MiekgDNSFields decodes many strings encoded using the miekg/dns module. It
// claims to be RFC1035 compliant.  However I disagree with how it handles escaped quotes.
// Input:             Output:
// `foo`               []string {`foo`}
// `one two`           []string {`one`, `two`}
// `one "two"`         []string {`one`, `two`}
// `"foo"`             []string {`foo`}
// `"foo" "bar"`       []string {`foo`, `bar`}
// `"es\"caped"`       []string {`es\"caped`}
// `"bumble\bee"`      []string {`bumble\bee`}
// `""doublequoted""`  []string {``, `doublequoted`, ``}
// `"\"escquoted\""`   []string {`\"escquoted\"`}
// `"in"side"`         error
// `"do""ble"`         []string {`do`, `ble`}
func MiekgDNSFields(s string) ([]string, error) {
	rr, err := dns.NewRR("example.com. IN TXT " + s)
	if err != nil {
		return nil, fmt.Errorf("could not parse %q TXT: %w", s, err)
	}

	return rr.(*dns.TXT).Txt, nil
}

// RFC1035Fields decodes many strings encoded using Tom's interpretation of RFC1035.
// Input:             Output:
// `foo`               []string {`foo`}
// `one two`           []string {`one`, `two`}
// `one "two"`         []string {`one`, `two`}
// `"foo"`             []string {`foo`}
// `"foo" "bar"`       []string {`foo`, `bar`}
// `"es\"caped"`       []string {`es"caped`}
// `"bumble\bee"`      []string {`bumble\bee`}
// `""doublequoted""`  []string {``, `doublequoted`, ``}
// `"\"escquoted\""`   []string {`"escquoted"`}
// `"in"side"`         error
// `"do""ble"`         []string {`do`, `ble`}

func RFC1035Fields(s string) ([]string, error) {
	return deEscape(MiekgDNSFields(s))
	// NB(tlim)I disagree with the miekg/dns parsing algorithm but it
	// seems to work.  We might replace it in the future.
}

func deEscape(sl []string, err error) ([]string, error) {
	if err != nil {
		return sl, err
	}
	for i := range sl {
		sl[i] = strings.ReplaceAll(sl[i], `\"`, `"`)
	}
	return sl, nil
}
