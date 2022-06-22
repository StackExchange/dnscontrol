package decode

import (
	"fmt"
	"reflect"
	"testing"
)

type testData struct {
	i string
	o [][]string
}

var tests = []testData{
	//  INPUT           QUOTED    ESCAPED    MIEKG/DNS   RFC1035
	{`foo`, [][]string{{`ERROR`}, {`ERROR`}, {`foo`}, {`foo`}}},
	{`one two`, [][]string{{`ERROR`}, {`ERROR`}, {`one`, `two`}, {`one`, `two`}}},
	{`one "two"`, [][]string{{`ERROR`}, {`ERROR`}, {`one`, `two`}, {`one`, `two`}}},
	{`"foo"`, [][]string{{`foo`}, {`foo`}, {`foo`}, {`foo`}}},
	{`"foo" "bar"`, [][]string{{`foo`, `bar`}, {`foo`, `bar`}, {`foo`, `bar`}, {`foo`, `bar`}}},
	{`"foo bar"`, [][]string{{`foo bar`}, {`foo bar`}, {`foo bar`}, {`foo bar`}}},
	{`"es\"caped"`, [][]string{{`es\"caped`}, {`es"caped`}, {`es\"caped`}, {`es"caped`}}},
	{`"bumble\bee"`, [][]string{{`bumble\bee`}, {`bumble\bee`}, {`bumble\bee`}, {`bumble\bee`}}},
	{`""doublequoted""`, [][]string{{`"doublequoted"`}, {`UNDEF`}, {``, `doublequoted`, ``}, {``, `doublequoted`, ``}}},
	{`"\"escquoted\""`, [][]string{{`\"escquoted\"`}, {`"escquoted"`}, {`\"escquoted\"`}, {`"escquoted"`}}},
	{`"in"side"`, [][]string{{`UNDEF`}, {`UNDEF`}, {`ERROR`}, {`ERROR`}}},
	{`"do""ble"`, [][]string{{`UNDEF`}, {`UNDEF`}, {`do`, `ble`}, {`do`, `ble`}}},
	{`1 2 3`, [][]string{{`ERROR`}, {`ERROR`}, {`1`, `2`, `3`}, {`1`, `2`, `3`}}},
	{`1 "2" 3`, [][]string{{`ERROR`}, {`ERROR`}, {`1`, `2`, `3`}, {`1`, `2`, `3`}}},
	{`1 "two" 3`, [][]string{{`ERROR`}, {`ERROR`}, {`1`, `two`, `3`}, {`1`, `two`, `3`}}},
	{`1 2 "qu\"te" "4"`, [][]string{{`ERROR`}, {`ERROR`}, {`1`, `2`, `qu\"te`, `4`}, {`1`, `2`, `qu"te`, `4`}}},
	//
	{`0 issue "letsencrypt.org; validationmethods=dns-01; accounturi=https://acme-v02.api.letsencrypt.org/acme/acct/1234"`,
		[][]string{
			{`ERROR`}, // QUOTED
			{`ERROR`}, // ESCAPED
			{`0`, `issue`, `letsencrypt.org; validationmethods=dns-01; accounturi=https://acme-v02.api.letsencrypt.org/acme/acct/1234`}, // MIEKG
			{`0`, `issue`, `letsencrypt.org; validationmethods=dns-01; accounturi=https://acme-v02.api.letsencrypt.org/acme/acct/1234`}, // RFC1035
		}},
}

// TODO(tlim): Is there a better way to represent this test data?  Is
// there a way to not have to manually update the comments in
// decoders.go?

func TestQuotedFields(t *testing.T) {
	parserTest(t, 0, QuotedFields, "QuotedFields")
}

func TestQuoteEscapedFields(t *testing.T) {
	parserTest(t, 1, QuoteEscapedFields, "QuoteEscapedFields")
}

func TestMiekgDNSFields(t *testing.T) {
	parserTest(t, 2, MiekgDNSFields, "miekgDNSFields")
}

func TestRFC1035Fields(t *testing.T) {
	parserTest(t, 3, RFC1035Fields, "RFC1035Fields")
}

func parserTest(t *testing.T, dataIndex int, f func(s string) ([]string, error), name string) {
	t.Helper()

	for i, tt := range tests {
		s := tt.i
		want := tt.o[dataIndex]
		first := tt.o[dataIndex][0]
		wantErr := (first == `ERROR`)
		if first == `UNDEF` {
			continue
		}
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			got, err := f(s)
			if wantErr {
				// We wanted an error, but didn't get one.
				if err == nil {
					t.Errorf("%s() expected error. Got none: (%v)", name, s)
					return
				}
				return
			}
			if err != nil {
				// We didn't want an error, but got one.
				t.Errorf("%s() unexpected error err=%v (%v)", name, err, s)
				return
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%s() = %v, want %v", name, got, want)
			}
		})
	}
}
