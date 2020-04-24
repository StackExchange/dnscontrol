package models

import "testing"

func TestHasRecordTypeName(t *testing.T) {
	x := &RecordConfig{
		Type: "A",
		Name: "@",
	}
	recs := Records{}
	if recs.HasRecordTypeName("A", "@") {
		t.Errorf("%v: expected (%v) got (%v)\n", recs, false, true)
	}
	recs = append(recs, x)
	if !recs.HasRecordTypeName("A", "@") {
		t.Errorf("%v: expected (%v) got (%v)\n", recs, true, false)
	}
	if recs.HasRecordTypeName("AAAA", "@") {
		t.Errorf("%v: expected (%v) got (%v)\n", recs, false, true)
	}
}

func TestKey(t *testing.T) {
	var tests = []struct {
		rc       RecordConfig
		expected RecordKey
	}{
		{
			RecordConfig{Type: "A", NameFQDN: "example.com"},
			RecordKey{Type: "A", NameFQDN: "example.com"},
		},
		{
			RecordConfig{Type: "R53_ALIAS", NameFQDN: "example.com"},
			RecordKey{Type: "R53_ALIAS", NameFQDN: "example.com"},
		},
		{
			RecordConfig{Type: "R53_ALIAS", NameFQDN: "example.com", R53Alias: map[string]string{"foo": "bar"}},
			RecordKey{Type: "R53_ALIAS", NameFQDN: "example.com"},
		},
		{
			RecordConfig{Type: "R53_ALIAS", NameFQDN: "example.com", R53Alias: map[string]string{"type": "AAAA"}},
			RecordKey{Type: "R53_ALIAS_AAAA", NameFQDN: "example.com"},
		},
	}
	for i, test := range tests {
		actual := test.rc.Key()
		if test.expected != actual {
			t.Errorf("%d: Expected %s, got %s", i, test.expected, actual)
		}
	}
}
