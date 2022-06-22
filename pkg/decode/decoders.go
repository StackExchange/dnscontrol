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
// Input:          Output:
// `foo`           error
// `"foo"`         []string{`foo`)
// `"foo" "bar"`   []string{`foo`, `bar`)
// `"fo\"o"`       undefined
// `"fo"o"`        undefined
// `"foo""bar"`    undefined
// `"foo\bar"`     undefined
// `""quoted""``   []string{'"quoted"`}
// `"\"quoted\""`` []string{'\"quoted\"`}
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
// Input:          Output:
// `foo`           error
// `"foo"`         []string{`foo`)
// `"foo" "bar"`   []string{`foo`, `bar`)
// `"fo\"o"`       []string{`fo"o`}
// `"fo"o"`        undefined
// `"foo""bar"`    undefined
// `"foo\bar"`     []string{`foobar`}
// `""quoted""``   undefined
// `"\"quoted\""`` []string{'"quoted"`}
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

func RFC1035Fields(s string) ([]string, error) {
	return miekgDNSFields(s)
	// NB(tlim)I disagree with the miekg/dns parsing algorithm but it
	// seems to work.  We might replace it in the future.
}

// miekgDNSFields decodes many strings encoded using the miekg/dns module. It
// claims to be RFC1035 compliant.  It accepts fields that aren't
// quoted.  It replaces escaped quotes with... escaped quotes; which
// round-trips properly but can't possibly be valid.
// Input:          Output:
// `foo`           error
// `"foo"`         []string{`foo`)
// `"foo" "bar"`   []string{`foo`, `bar`)
// `"fo\"o"`       []string{`fo"o`}
// `"fo"o"`        undefined
// `"foo""bar"`    undefined
// `"foo\bar"`     []string{`foobar`}
// `""quoted""``   undefined
// `"\"quoted\""`` []string{`"quoted"``}
func miekgDNSFields(s string) ([]string, error) {
	rr, err := dns.NewRR("example.com. IN TXT " + s)
	if err != nil {
		return nil, fmt.Errorf("could not parse %q TXT: %w", s, err)
	}

	return rr.(*dns.TXT).Txt, nil
}
