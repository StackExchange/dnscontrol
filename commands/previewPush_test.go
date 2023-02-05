package commands

import (
	"github.com/StackExchange/dnscontrol/v3/pkg/credsfile"
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
			gotReplacementType, gotWarnMsg, err := refineProviderType("foo", tt.args.t, tt.args.credFields, "FOO")
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

func TestIsDomainOrZoneManaged(t *testing.T) {
	DNSConfigArgs := GetDNSConfigArgs{
		ExecuteDSLArgs: ExecuteDSLArgs{
			JSFile: "test_data/dnsconfig.js",
		},
	}

	cfg, err := GetDNSConfig(DNSConfigArgs)
	if err != nil {
		t.Errorf("failed getting dns config. err: %s", err)
	}

	providerConfigs, err := credsfile.LoadProviderConfigs("test_data/bind-creds.json")
	if err != nil {
		t.Errorf("failed loading provider config. err: %s", err)
	}

	_, _, _, err = InitializeProviders(cfg, providerConfigs, false)

	if err != nil {
		t.Errorf("error initializing providers. err: %s", err)
	}
	tests := []struct {
		name          string
		domainOrZone  string
		providerName  string
		domain        string
		wantIsManaged bool
	}{
		{"domain/0", "domain", "registrar1", "example.org", true},
		{"domain/1", "domain", "registrar1", "example.com", false},
		{"domain/2", "domain", "registrar2", "example.org", false},
		{"domain/3", "domain", "registrar2", "example.com", true},

		{"zone/0", "zone", "dsp1", "example.org", true},
		{"zone/1", "zone", "dsp1", "example.com", false},
		{"zone/2", "zone", "dsp2", "example.org", false},
		{"zone/3", "zone", "dsp2", "example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.domainOrZone == "domain" {
				gotIsManaged := IsDomainManagedByRegistrar(cfg, tt.domain, tt.providerName)
				if gotIsManaged != tt.wantIsManaged {
					t.Errorf("IsDomainManagedByRegistrar() gotIsManaged = %v, want %v", gotIsManaged, tt.wantIsManaged)
				}

			}
			if tt.domainOrZone == "zone" {
				gotIsManaged := IsZoneManagedByProvider(cfg, tt.domain, tt.providerName)
				if gotIsManaged != tt.wantIsManaged {
					t.Errorf("IsZoneManagedByProvider() gotIsManaged = %v, want %v", gotIsManaged, tt.wantIsManaged)
				}
			}
		})
	}

}
