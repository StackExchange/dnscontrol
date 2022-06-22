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
	//  INPUT           QUOTED    ESCAPED    MIEKG/DNS
	{`foo`, [][]string{{`ERROR`}, {`ERROR`}, {`ERROR`}}},
	{`"foo"`, [][]string{{`foo`}, {`ERROR`}, {`ERROR`}}},
	{`"foo" "bar"`, [][]string{{`foo`, `bar`}, {`ERROR`}, {`ERROR`}}},
	{`"es\"caped"`, [][]string{{`es\"caped`}, {`ERROR`}, {`ERROR`}}},
	{`"in"side"`, [][]string{{`UNDEF`}, {`ERROR`}, {`ERROR`}}},
	{`"do""ble"`, [][]string{{`UNDEF`}, {`ERROR`}, {`ERROR`}}},
	{`"bumble\bee"`, [][]string{{`bumble\bee`}, {`ERROR`}, {`ERROR`}}},
	{`""quoted""`, [][]string{{`"quoted"`}, {`ERROR`}, {`ERROR`}}},
	{`"\"escquoted\""`, [][]string{{`\"escquoted\"`}, {`ERROR`}, {`ERROR`}}},
}

func TestQuotedFields(t *testing.T) {
	const SELECT = 0
	for i, tt := range tests {
		s := tt.i
		want := tt.o[SELECT]
		first := tt.o[SELECT][0]
		wantErr := (first == `ERROR`)
		if first == `UNDEF` {
			continue
		}
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			got, err := QuotedFields(s)
			if wantErr {
				// We wanted an error, but didn't get one.
				if err == nil {
					t.Errorf("DecodeQuotedFields() expected error. Got none: (%v)", s)
					return
				}
				return
			}
			if err != nil {
				// We didn't want an error, but got one.
				t.Errorf("DecodeQuotedFields() unexpected error err=%v (%v)", err, s)
				return
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("DecodeQuotedFields() = %v, want %v", got, want)
			}
		})
	}
}
