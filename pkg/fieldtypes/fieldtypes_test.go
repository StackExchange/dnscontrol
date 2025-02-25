package fieldtypes

import "testing"

func TestParseLabel3(t *testing.T) {
	type args struct {
		short     string
		subdomain string
		origin    string
	}
	tests := []struct {
		name    string
		args    args
		short   string
		fqdn    string
		wantErr bool
	}{
		// D_EXTEND() mode (subdomain)
		{"sd null", args{"", "subdomain", "origin"}, "subdomain", "subdomain.origin", false},
		{"sd apex", args{"@", "subdomain", "origin"}, "subdomain", "subdomain.origin", false},
		{"sd dot", args{".", "subdomain", "origin"}, "", "", true},
		{"sd normal", args{"short", "subdomain", "origin"}, "short.subdomain", "short.subdomain.origin", false},
		{"sd dot err 0", args{"short.", "subdomain", "origin"}, "", "", true},
		{"sd dot err 1", args{"foo.short.", "subdomain", "origin"}, "", "", true},
		{"sd dot apex", args{"origin.", "subdomain", "origin"}, "@", "origin", false},
		{"sd dot 1", args{"short.origin.", "subdomain", "origin"}, "short", "short.origin", false},
		{"sd dot 2", args{"foo.short.origin.", "subdomain", "origin"}, "foo.short", "foo.short.origin", false},
		// D() mode (no subdomain)
		{"null", args{"", "", "origin"}, "@", "origin", false},
		{"apex", args{"@", "", "origin"}, "@", "origin", false},
		{"dot", args{".", "", "origin"}, "", "", true},
		{"normal", args{"short", "", "origin"}, "short", "short.origin", false},
		{"dot err 0", args{"short.", "", "origin"}, "", "", true},
		{"dot err 1", args{"foo.short.", "", "origin"}, "", "", true},
		{"dot apex", args{"origin.", "", "origin"}, "@", "origin", false},
		{"dot 1", args{"short.origin.", "", "origin"}, "short", "short.origin", false},
		{"dot 2", args{"foo.short.origin.", "", "origin"}, "foo.short", "foo.short.origin", false},

		// Legacy mode (no origin)

		// D_EXTEND() mode (subdomain)
		{"leg sd null", args{"", "subdomain", ""}, "subdomain", "", false},
		{"leg sd apex", args{"@", "subdomain", ""}, "subdomain", "", false},
		{"leg sd dot", args{".", "subdomain", ""}, "", "", true},
		{"leg sd normal", args{"short", "subdomain", ""}, "short.subdomain", "", false},
		//{"leg sd dot err 0", args{"short.", "subdomain", ""}, "", "", true}, // Test depends on the origin
		//{"leg sd dot err 1", args{"foo.short.", "subdomain", ""}, "", "", true}, // Test depends on the origin
		{"leg sd dot apex", args{"origin.", "subdomain", ""}, "", "", true},
		{"leg sd dot 1", args{"short.origin.", "subdomain", ""}, "", "", true},
		{"leg sd dot 2", args{"foo.short.origin.", "subdomain", ""}, "", "", true},
		// D() mode (no subdomain)
		{"leg null", args{"", "", ""}, "@", "", false},
		{"leg apex", args{"@", "", ""}, "@", "", false},
		{"leg dot", args{".", "", ""}, "", "", true},
		{"leg normal", args{"short", "", ""}, "short", "", false},
		{"leg dot err 0", args{"short.", "", ""}, "", "", true},
		{"leg dot err 1", args{"foo.short.", "", ""}, "", "", true},
		{"leg dot apex", args{"origin.", "", ""}, "", "", true},
		{"leg dot 1", args{"short.origin.", "", ""}, "", "", true},
		{"leg dot 2", args{"foo.short.origin.", "", ""}, "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLabel, gotFQDN, err := ParseLabel3(tt.args.short, tt.args.subdomain, tt.args.origin)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseLabel3() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else {
				if gotLabel != tt.short {
					t.Errorf("ParseLabel3() label = %q, want %q", gotLabel, tt.short)
				}
				if gotFQDN != tt.fqdn {
					t.Errorf("ParseLabel3() labelFQDN = %q, want %q", gotFQDN, tt.fqdn)
				}
			}
		})
	}
}
func TestParseHostnameDot(t *testing.T) {
	type args struct {
		short  string
		origin string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"origin with dot", args{"short", "subdomain.origin."}, "FAIL", true},
		{"origin uppercase", args{"short", "subdomain.Origin"}, "FAIL", true},
		{"subdomain uppercase", args{"short", "Subdomain.origin"}, "FAIL", true},
		{"empty short", args{"", "subdomain.origin"}, "FAIL", true},
		{"empty short no sub", args{"", "origin"}, "FAIL", true},
		{"double-dots", args{"foo..", "origin"}, "FAIL", true},

		{"short uppercase", args{"Short", "subdomain.origin"}, "short.subdomain.origin.", false},
		{"short with dot", args{"short.", "subdomain.origin"}, "short.", false},
		{"short is *", args{"*", "subdomain.origin"}, "*.subdomain.origin.", false},
		{"short is * no sub", args{"*", "origin"}, "*.origin.", false},
		{"subdomain in use", args{"short", "subdomain.origin"}, "short.subdomain.origin.", false},
		{"subdomain in use with @", args{"@", "subdomain.origin"}, "subdomain.origin.", false},
		{"short is @", args{"@", "origin"}, "origin.", false},
		{"normal case", args{"short", "origin"}, "short.origin.", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHostnameDot(tt.args.short, tt.args.origin)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseHostnameDot() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else {
				if got != tt.want {
					t.Errorf("ParseHostnameDot() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
