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
		{``, ``, []string{``}},
		{`foo`, `foo`, []string{`foo`}},
		{`"foo"`, `foo`, []string{`foo`}},
		{`"aaa" "bbb"`, `aaa`, []string{`aaa`, `bbb`}},
	}
	for i, test := range tests {
		x := &RecordConfig{Type: "TXT"}
		x.SetTxtParse(test.d1)
		if x.Target != test.e1 {
			t.Errorf("%v: expected Target=(%v) got (%v)", i, test.e1, x.Target)
		}
	}
}
