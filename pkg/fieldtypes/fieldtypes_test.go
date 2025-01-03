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
		{"sd null", args{"", "subdomain", "origin"}, "subdomain", "subdomain.origin", false},
		{"sd apex", args{"@", "subdomain", "origin"}, "subdomain", "subdomain.origin", false},
		{"sd dot", args{".", "subdomain", "origin"}, "", "", true},
		{"sd normal", args{"short", "subdomain", "origin"}, "short.subdomain", "short.subdomain.origin", false},
		{"sd dot err 0", args{"short.", "subdomain", "origin"}, "", "", true},
		{"sd dot err 1", args{"foo.short.", "subdomain", "origin"}, "", "", true},
		{"sd dot apex", args{"origin.", "subdomain", "origin"}, "@", "origin", false},
		{"sd dot 1", args{"short.origin.", "subdomain", "origin"}, "short", "short.origin", false},
		{"sd dot 2", args{"foo.short.origin.", "subdomain", "origin"}, "foo.short", "foo.short.origin", false},
		//
		{"null", args{"", "", "origin"}, "@", "origin", false},
		{"apex", args{"@", "", "origin"}, "@", "origin", false},
		{"dot", args{".", "", "origin"}, "", "", true},
		{"normal", args{"short", "", "origin"}, "short", "short.origin", false},
		{"dot err 0", args{"short.", "", "origin"}, "", "", true},
		{"dot err 1", args{"foo.short.", "", "origin"}, "", "", true},
		{"dot apex", args{"origin.", "", "origin"}, "@", "origin", false},
		{"dot 1", args{"short.origin.", "", "origin"}, "short", "short.origin", false},
		{"dot 2", args{"foo.short.origin.", "", "origin"}, "foo.short", "foo.short.origin", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLabel, gotFQDN, err := ParseLabel3(tt.args.short, tt.args.subdomain, tt.args.origin)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLabel3() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLabel != tt.short {
				t.Errorf("ParseLabel3() label = %v, want %v", gotLabel, tt.short)
			}
			if gotFQDN != tt.fqdn {
				t.Errorf("ParseLabel3() labelFQDN = %v, want %v", gotFQDN, tt.fqdn)
			}
		})
	}
}
