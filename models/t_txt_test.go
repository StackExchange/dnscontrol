package models

import (
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
		r := validateTXT(rc)
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
		r := validateTXT(rc)
		if test.e1 != (r == nil) {
			s := "<nil>"
			if r == nil {
				s = "error"
			}
			t.Errorf("%v: expected (%v) got: (%v)", i, s, r)
		}
	}
}
