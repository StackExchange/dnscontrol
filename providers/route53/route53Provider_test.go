package route53

import "testing"

func TestUnescape(t *testing.T) {
	var tests = []struct {
		experiment, expected string
	}{
		{"foo", "foo"},
		{"foo.", "foo"},
		{"foo..", "foo."},
		{"foo...", "foo.."},
		{`\052`, "*"},
		{`\052.foo..`, "*.foo."},
		// {`\053.foo`, "+.foo"},  // Not implemented yet.
	}

	for i, test := range tests {
		actual := unescape(&test.experiment)
		if test.expected != actual {
			t.Errorf("%d: Expected %s, got %s", i, test.expected, actual)
		}
	}
}
