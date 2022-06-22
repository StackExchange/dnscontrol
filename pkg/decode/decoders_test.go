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
	{`"es\"caped"`, [][]string{{`es\"caped`}, {`es"caped`}, {`es\"caped`}, {`es"caped`}}},
	{`"bumble\bee"`, [][]string{{`bumble\bee`}, {`bumble\bee`}, {`bumble\bee`}, {`bumble\bee`}}},
	{`""doublequoted""`, [][]string{{`"doublequoted"`}, {`UNDEF`}, {``, `doublequoted`, ``}, {``, `doublequoted`, ``}}},
	{`"\"escquoted\""`, [][]string{{`\"escquoted\"`}, {`"escquoted"`}, {`\"escquoted\"`}, {`"escquoted"`}}},
	{`"in"side"`, [][]string{{`UNDEF`}, {`UNDEF`}, {`ERROR`}, {`ERROR`}}},
	{`"do""ble"`, [][]string{{`UNDEF`}, {`UNDEF`}, {`do`, `ble`}, {`do`, `ble`}}},
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
