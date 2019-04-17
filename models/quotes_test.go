package models

import (
	"testing"
)

func TestIsQuoted(t *testing.T) {
	tests := []struct {
		d1 string
		e1 bool
	}{
		{``, false},
		{`foo`, false},
		{`""`, true},
		{`"a"`, true},
		{`"bb"`, true},
		{`"ccc"`, true},
		{`"aaa" "bbb"`, true},
	}
	for i, test := range tests {
		r := IsQuoted(test.d1)
		if r != test.e1 {
			t.Errorf("%v: expected (%v) got (%v)", i, test.e1, r)
		}
	}
}

func TestStripQuotes(t *testing.T) {
	tests := []struct {
		d1 string
		e1 string
	}{
		{``, ``},
		{`a`, `a`},
		{`bb`, `bb`},
		{`ccc`, `ccc`},
		{`dddd`, `dddd`},
		{`"A"`, `A`},
		{`"BB"`, `BB`},
		{`"CCC"`, `CCC`},
		{`"DDDD"`, `DDDD`},
		{`"EEEEE"`, `EEEEE`},
		{`"aaa" "bbb"`, `aaa" "bbb`},
	}
	for i, test := range tests {
		r := StripQuotes(test.d1)
		if r != test.e1 {
			t.Errorf("%v: expected (%v) got (%v)", i, test.e1, r)
		}
	}
}

func TestSetTxtParse(t *testing.T) {
	tests := []struct {
		d1 string
		e1 string
		e2 []string
	}{
		{`foo`, `foo`, []string{`foo`}},
		{`"foo"`, `foo`, []string{`foo`}},
		{`"foo bar"`, `foo bar`, []string{`foo bar`}},
		{`foo bar`, `foo bar`, []string{`foo bar`}},
		{`"aaa" "bbb"`, `aaa`, []string{`aaa`, `bbb`}},
	}
	for i, test := range tests {
		ls := ParseQuotedTxt(test.d1)
		if ls[0] != test.e1 {
			t.Errorf("%v: expected Target=(%v) got (%v)", i, test.e1, ls[0])
		}
		if len(ls) != len(test.e2) {
			t.Errorf("%v: expected TxtStrings=(%v) got (%v)", i, test.e2, ls)
		}
		for i := range ls {
			if len(ls[i]) != len(test.e2[i]) {
				t.Errorf("%v: expected TxtStrings=(%v) got (%v)", i, test.e2, ls)
			}
		}
	}
}
