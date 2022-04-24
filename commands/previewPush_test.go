package commands

import (
	"strings"
	"testing"
)

func Test_refineProviderType(t *testing.T) {

	var mapEmpty map[string]string
	mapTypeMissing := map[string]string{"otherfield": "othervalue"}
	mapTypeFoo := map[string]string{"TYPE": "FOO"}
	mapTypeBar := map[string]string{"TYPE": "BAR"}
	mapTypeHyphen := map[string]string{"TYPE": "-"}

	type args struct {
		t          string
		credFields map[string]string
	}
	tests := []struct {
		name                string
		args                args
		wantReplacementType string
		wantWarnMsgPrefix   string
		wantErr             bool
	}{
		{"fooEmp", args{"FOO", mapEmpty}, "FOO", "WARN", false},       // 3.x: Provide compatibility suggestion. 4.0: hard error
		{"fooMis", args{"FOO", mapTypeMissing}, "FOO", "WARN", false}, // 3.x: Provide compatibility suggestion. 4.0: hard error
		{"fooHyp", args{"FOO", mapTypeHyphen}, "-", "", true},         // Error: Invalid creds.json data.
		{"fooFoo", args{"FOO", mapTypeFoo}, "FOO", "INFO", false},     // Suggest cleanup.
		{"fooBar", args{"FOO", mapTypeBar}, "FOO", "", true},          // Error: Mismatched!

		{"hypEmp", args{"-", mapEmpty}, "", "", true},       // Hard error. creds.json entry is missing type.
		{"hypMis", args{"-", mapTypeMissing}, "", "", true}, // Hard error. creds.json entry is missing type.
		{"hypHyp", args{"-", mapTypeHyphen}, "-", "", true}, // Hard error: Invalid creds.json data.
		{"hypFoo", args{"-", mapTypeFoo}, "FOO", "", false}, // normal
		{"hypBar", args{"-", mapTypeBar}, "BAR", "", false}, // normal
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr && (tt.wantWarnMsgPrefix != "") {
				t.Error("refineProviderType() bad test data. Prefix should be \"\" if wantErr is set")
			}
			gotReplacementType, gotWarnMsg, err := refineProviderType("foo", tt.args.t, tt.args.credFields)
			if !strings.HasPrefix(gotWarnMsg, tt.wantWarnMsgPrefix) {
				t.Errorf("refineProviderType() gotWarnMsg = %q, wanted prefix %q", gotWarnMsg, tt.wantWarnMsgPrefix)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("refineProviderType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotReplacementType != tt.wantReplacementType {
				t.Errorf("refineProviderType() gotReplacementType = %q, want %q (warn,msg)=(%q,%s)", gotReplacementType, tt.wantReplacementType, gotWarnMsg, err)
			}
		})
	}
}
