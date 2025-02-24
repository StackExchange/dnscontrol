package rfc1035

import (
	"reflect"
	"testing"
)

func TestParseZonefileOperands(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"192.0.2.1",
			[]string{"192.0.2.1"}},
		{"10 mail.example.org.",
			[]string{"10", "mail.example.org."}},
		{`"v=spf1 include:example.net ~all"`,
			[]string{`v=spf1 include:example.net ~all`}},
		{`0 issue "letsencrypt.org; validationmethods=dns-01; accounturi=https://acme-v02.api.letsencrypt.org/acme/acct/1234"`,
			[]string{"0", "issue", `letsencrypt.org; validationmethods=dns-01; accounturi=https://acme-v02.api.letsencrypt.org/acme/acct/1234`}},
	}

	for _, test := range tests {
		result, err := Fields(test.input)
		if err != nil {
			t.Errorf("ParseZonefileOperands(%q) = %v; want %v", test.input, err, nil)
		}
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("ParseZonefileOperands(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}
