package models

import "testing"

func TestKey(t *testing.T) {
	var tests = []struct {
		rc       RecordConfig
		expected RecordKey
	}{
		{
			RecordConfig{Type: "A", Name: "@"},
			RecordKey{Type: "A", Name: "@"},
		},
		{
			RecordConfig{Type: "R53_ALIAS", Name: "@"},
			RecordKey{Type: "R53_ALIAS", Name: "@"},
		},
		{
			RecordConfig{Type: "R53_ALIAS", Name: "@", R53Alias: map[string]string{"foo": "bar"}},
			RecordKey{Type: "R53_ALIAS", Name: "@"},
		},
		{
			RecordConfig{Type: "R53_ALIAS", Name: "@", R53Alias: map[string]string{"type": "AAAA"}},
			RecordKey{Type: "R53_ALIAS_AAAA", Name: "@"},
		},
	}
	for i, test := range tests {
		actual := test.rc.Key()
		if test.expected != actual {
			t.Errorf("%d: Expected %s, got %s", i, test.expected, actual)
		}
	}
}
