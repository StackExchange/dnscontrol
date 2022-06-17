package cloudflare

import "testing"

func Test_isCloudflareQuoteBug(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{"01", args{`basic`}, false},
		{"02", args{`"one" "two"`}, false},
		{"03", args{`"eh bee"`}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCloudflareQuoteBug(tt.args.s); got != tt.want {
				t.Errorf("isCloudflareQuoteBug() = %v, want %v (%s)", got, tt.want, tt.args.s)
			}
		})
	}
}

func Test_fixCloudflareQuoteBug(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"03", args{`"eh bee"`}, `"\"eh bee\""`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixCloudflareQuoteBug(tt.args.s); got != tt.want {
				t.Errorf("fixCloudflareQuoteBug() = %q, want %q", got, tt.want)
			}
		})
	}
}
