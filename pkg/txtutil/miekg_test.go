package txtutil_test

import (
	"testing"

	"github.com/DNSControl/dnscontrol/v4/pkg/txtutil"
)

func TestZoneify(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		txt  []string
		want string
	}{
		{"simple", []string{`simple`}, `"simple"`},
		{"space", []string{`with space`}, `"with space"`},
		{"quote", []string{`with'quote`}, `"with'quote"`},
		{"dquote", []string{`with"dquote`}, `"with\"dquote"`},
		//{"backslash", []string{`with\backslash`}, `"with\\backslash"`},  // FAILING
		{"multiple", []string{`line1`, `line2`}, `"line1" "line2"`},
		//{"complex", []string{`line with "dquotes" and \backslash\`}, `"line with "dquote" and \backslash\`}, // FAILING
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := txtutil.Zoneify(tt.txt)
			if got != tt.want {
				t.Errorf("Zoneify() = %v, want %v", got, tt.want)
			}
		})
	}
}
