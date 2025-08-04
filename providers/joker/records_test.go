package joker

import (
	"reflect"
	"testing"
)

func TestParseZoneLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple A record",
			input:    "www A 0 1.2.3.4 300",
			expected: []string{"www", "A", "0", "1.2.3.4", "300"},
		},
		{
			name:     "root domain A record",
			input:    "@ A 0 127.0.0.1 300",
			expected: []string{"@", "A", "0", "127.0.0.1", "300"},
		},
		{
			name:     "CNAME record",
			input:    "mail CNAME 0 example.com. 300",
			expected: []string{"mail", "CNAME", "0", "example.com.", "300"},
		},
		{
			name:     "NS record for subdomain",
			input:    "sub NS 0 ns1.example.com. 300",
			expected: []string{"sub", "NS", "0", "ns1.example.com.", "300"},
		},
		{
			name:     "NS record multiple nameservers",
			input:    "delegate NS 0 ns.provider.com. 3600",
			expected: []string{"delegate", "NS", "0", "ns.provider.com.", "3600"},
		},
		{
			name:     "MX record with priority",
			input:    "@ MX 10 mail.example.com. 300",
			expected: []string{"@", "MX", "10", "mail.example.com.", "300"},
		},
		{
			name:     "TXT record with quoted content",
			input:    "@ TXT 0 \"v=spf1 include:_spf.google.com ~all\" 300",
			expected: []string{"@", "TXT", "0", "\"v=spf1 include:_spf.google.com ~all\"", "300"},
		},
		{
			name:     "TXT record with spaces in quotes",
			input:    "test TXT 0 \"this has multiple spaces\" 300",
			expected: []string{"test", "TXT", "0", "\"this has multiple spaces\"", "300"},
		},
		{
			name:     "TXT record with escaped quotes",
			input:    "test TXT 0 \"value with \\\"escaped\\\" quotes\" 300",
			expected: []string{"test", "TXT", "0", "\"value with \\\"escaped\\\" quotes\"", "300"},
		},
		{
			name:     "SRV record",
			input:    "_sip._tcp SRV 10/20 sip.example.com:5060 300",
			expected: []string{"_sip._tcp", "SRV", "10/20", "sip.example.com:5060", "300"},
		},
		{
			name:     "CAA record",
			input:    "@ CAA 0 0 issue \"letsencrypt.org\" 300",
			expected: []string{"@", "CAA", "0", "0", "issue", "\"letsencrypt.org\"", "300"},
		},
		{
			name:     "NAPTR record",
			input:    "@ NAPTR 100/50 target.example.com. \"u\" \"sip+E2U\" \"!^.*$!sip:info@example.com!\" 300",
			expected: []string{"@", "NAPTR", "100/50", "target.example.com.", "\"u\"", "\"sip+E2U\"", "\"!^.*$!sip:info@example.com!\"", "300"},
		},
		{
			name:     "multiple consecutive spaces",
			input:    "www    A    0    1.2.3.4    300",
			expected: []string{"www", "A", "0", "1.2.3.4", "300"},
		},
		{
			name:     "tabs and spaces mixed",
			input:    "www\tA\t0\t1.2.3.4\t300",
			expected: []string{"www", "A", "0", "1.2.3.4", "300"},
		},
		{
			name:     "quoted content with tabs and spaces",
			input:    "test TXT 0 \"content\twith\ttabs and   spaces\" 300",
			expected: []string{"test", "TXT", "0", "\"content\twith\ttabs and   spaces\"", "300"},
		},
		{
			name:     "escaped backslash",
			input:    "test TXT 0 \"path\\\\to\\\\file\" 300",
			expected: []string{"test", "TXT", "0", "\"path\\\\to\\\\file\"", "300"},
		},
		{
			name:     "empty quoted content",
			input:    "test TXT 0 \"\" 300",
			expected: []string{"test", "TXT", "0", "\"\"", "300"},
		},
		{
			name:     "AAAA record",
			input:    "www AAAA 0 2001:db8::1 300",
			expected: []string{"www", "AAAA", "0", "2001:db8::1", "300"},
		},
		{
			name:     "subdomain with multiple dots",
			input:    "deep.sub.domain A 0 1.2.3.4 300",
			expected: []string{"deep.sub.domain", "A", "0", "1.2.3.4", "300"},
		},
		{
			name:     "TXT record with special characters",
			input:    "test TXT 0 \"key=value;tag=test\" 300",
			expected: []string{"test", "TXT", "0", "\"key=value;tag=test\"", "300"},
		},
		{
			name:     "short record without TTL",
			input:    "www A 0 1.2.3.4",
			expected: []string{"www", "A", "0", "1.2.3.4"},
		},
		{
			name:     "leading and trailing spaces",
			input:    "  www A 0 1.2.3.4 300  ",
			expected: []string{"www", "A", "0", "1.2.3.4", "300"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseZoneLine(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseZoneLine() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseZoneLineEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []string
		expectEmpty bool
	}{
		{
			name:        "empty string",
			input:       "",
			expectEmpty: true,
		},
		{
			name:        "only spaces",
			input:       "   ",
			expectEmpty: true,
		},
		{
			name:        "only tabs",
			input:       "\t\t\t",
			expectEmpty: true,
		},
		{
			name:     "single word",
			input:    "word",
			expected: []string{"word"},
		},
		{
			name:     "unmatched quote at start",
			input:    "\"unclosed quote",
			expected: []string{"\"unclosed quote"},
		},
		{
			name:     "unmatched quote at end",
			input:    "unclosed quote\"",
			expected: []string{"unclosed", "quote\""},
		},
		{
			name:     "consecutive quotes",
			input:    "test \"\" \"value\" more",
			expected: []string{"test", "\"\"", "\"value\"", "more"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseZoneLine(tt.input)

			if tt.expectEmpty {
				if len(got) != 0 {
					t.Errorf("parseZoneLine() = %v, expected empty slice", got)
				}
			} else {
				if !reflect.DeepEqual(got, tt.expected) {
					t.Errorf("parseZoneLine() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}
