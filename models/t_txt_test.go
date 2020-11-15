package models

import (
	"reflect"
	"strings"
	"testing"
)

func TestValidateTXT_single(t *testing.T) {
	tests := []struct {
		d1 string
		e1 bool
	}{
		{`x`, true},
		{`foo`, true},
		{strings.Repeat(`A`, 253), true},
		{strings.Repeat(`B`, 254), true},
		{strings.Repeat(`C`, 255), true},
		{strings.Repeat(`D`, 256), false},
		{strings.Repeat(`E`, 257), false},
	}
	for i, test := range tests {
		rc := &RecordConfig{Type: "TXT"}
		rc.SetTargetTXT(test.d1)
		r := ValidateTXT(rc)
		if test.e1 != (r == nil) {
			s := "<nil>"
			if r == nil {
				s = "error"
			}
			t.Errorf("%v: expected (%v) got (%v) (%q)", i, s, r, test.d1)
		}
	}
}

func TestValidateTXT_multi(t *testing.T) {
	s1 := "A"
	s253 := strings.Repeat("C", 253)
	s254 := strings.Repeat("C", 254)
	s255 := strings.Repeat("D", 255)
	s256 := strings.Repeat("E", 256)
	s257 := strings.Repeat("E", 256)
	tests := []struct {
		d1 []string
		e1 bool
	}{
		{[]string{`x`}, true},
		{[]string{`foo`}, true},
		{[]string{s253}, true},
		{[]string{s254}, true},
		{[]string{s255}, true},
		{[]string{s256}, false},
		{[]string{s257}, false},

		{[]string{s1, s254}, true},
		{[]string{s1, s255}, true},
		{[]string{s1, s256}, false},

		{[]string{s1, s254, s1}, true},
		{[]string{s1, s255, s1}, true},
		{[]string{s1, s256, s1}, false},
	}
	for i, test := range tests {
		rc := &RecordConfig{Type: "TXT"}
		rc.SetTargetTXTs(test.d1)
		r := ValidateTXT(rc)
		if test.e1 != (r == nil) {
			s := "<nil>"
			if r == nil {
				s = "error"
			}
			t.Errorf("%v: expected (%v) got: (%v)", i, s, r)
		}
	}
}

func Test_splitChunks(t *testing.T) {
	type args struct {
		buf string
		lim int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"0", args{"", 3}, []string{}},
		{"1", args{"a", 3}, []string{"a"}},
		{"2", args{"ab", 3}, []string{"ab"}},
		{"3", args{"abc", 3}, []string{"abc"}},
		{"4", args{"abcd", 3}, []string{"abc", "d"}},
		{"5", args{"abcde", 3}, []string{"abc", "de"}},
		{"6", args{"abcdef", 3}, []string{"abc", "def"}},
		{"7", args{"abcdefg", 3}, []string{"abc", "def", "g"}},
		{"8", args{"abcdefgh", 3}, []string{"abc", "def", "gh"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitChunks(tt.args.buf, tt.args.lim); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitChunks() = %v, want %v", got, tt.want)
			}
		})
	}
}
