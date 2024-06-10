package cloudflare

import "testing"

func Test_makeSingleDirectRule(t *testing.T) {
	tests := []struct {
		name string
		//
		pattern string
		replace string
		//
		wantMatch string
		wantExpr  string
		wantErr   bool
	}{
		{
			name:      "001",
			pattern:   "example.com/",
			replace:   "foo",
			wantMatch: `http.host eq "example.com" and http.request.uri.path eq "/"`,
			wantExpr:  `concat("https://example.com", http.request.uri.path)`,
			wantErr:   false,
		},
		// TODO: Add test cases from dnsconfig.js
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, gotExpr, err := makeRuleFromPattern(tt.pattern, tt.replace, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeSingleDirectRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMatch != tt.wantMatch {
				t.Errorf("makeSingleDirectRule() = %v, want %v", gotMatch, tt.wantMatch)
			}
			if gotExpr != tt.wantExpr {
				t.Errorf("makeSingleDirectRule() = %v, want %v", gotExpr, tt.wantExpr)
			}
		})
	}
}
