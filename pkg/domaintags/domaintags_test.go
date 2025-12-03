package domaintags

import (
	"testing"
)

func Test_MakeDomainFixForms(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantTag         string
		wantNameRaw     string
		wantNameIDN     string
		wantNameUnicode string
		wantUniqueName  string
		wantHasBang     bool
	}{
		{
			name:            "simple domain",
			input:           "example.com",
			wantTag:         "",
			wantNameRaw:     "example.com",
			wantNameIDN:     "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com",
			wantHasBang:     false,
		},
		{
			name:            "domain with tag",
			input:           "example.com!mytag",
			wantTag:         "mytag",
			wantNameRaw:     "example.com",
			wantNameIDN:     "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com!mytag",
			wantHasBang:     true,
		},
		{
			name:            "domain with empty tag",
			input:           "example.com!",
			wantTag:         "",
			wantNameRaw:     "example.com",
			wantNameIDN:     "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com!",
			wantHasBang:     true,
		},
		{
			name:            "unicode domain",
			input:           "उदाहरण.com",
			wantTag:         "",
			wantNameRaw:     "उदाहरण.com",
			wantNameIDN:     "xn--p1b6ci4b4b3a.com",
			wantNameUnicode: "उदाहरण.com",
			wantUniqueName:  "xn--p1b6ci4b4b3a.com",
			wantHasBang:     false,
		},
		{
			name:            "unicode domain with tag",
			input:           "उदाहरण.com!mytag",
			wantTag:         "mytag",
			wantNameRaw:     "उदाहरण.com",
			wantNameIDN:     "xn--p1b6ci4b4b3a.com",
			wantNameUnicode: "उदाहरण.com",
			wantUniqueName:  "xn--p1b6ci4b4b3a.com!mytag",
			wantHasBang:     true,
		},
		{
			name:            "punycode domain",
			input:           "xn--p1b6ci4b4b3a.com",
			wantTag:         "",
			wantNameRaw:     "xn--p1b6ci4b4b3a.com",
			wantNameIDN:     "xn--p1b6ci4b4b3a.com",
			wantNameUnicode: "उदाहरण.com",
			wantUniqueName:  "xn--p1b6ci4b4b3a.com",
			wantHasBang:     false,
		},
		{
			name:            "punycode domain with tag",
			input:           "xn--p1b6ci4b4b3a.com!mytag",
			wantTag:         "mytag",
			wantNameRaw:     "xn--p1b6ci4b4b3a.com",
			wantNameIDN:     "xn--p1b6ci4b4b3a.com",
			wantNameUnicode: "उदाहरण.com",
			wantUniqueName:  "xn--p1b6ci4b4b3a.com!mytag",
			wantHasBang:     true,
		},
		{
			name:            "mixed case domain",
			input:           "Example.COM",
			wantTag:         "",
			wantNameRaw:     "example.com",
			wantNameIDN:     "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com",
			wantHasBang:     false,
		},
		{
			name:            "mixed case domain with tag",
			input:           "Example.COM!MyTag",
			wantTag:         "MyTag",
			wantNameRaw:     "example.com",
			wantNameIDN:     "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com!MyTag",
			wantHasBang:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeDomainFixForms(tt.input)
			if got.Tag != tt.wantTag {
				t.Errorf("MakeDomainFixForms() gotTag = %v, want %v", got.Tag, tt.wantTag)
			}
			if got.NameRaw != tt.wantNameRaw {
				t.Errorf("MakeDomainFixForms() gotNameRaw = %v, want %v", got.NameRaw, tt.wantNameRaw)
			}
			if got.NameIDN != tt.wantNameIDN {
				t.Errorf("MakeDomainFixForms() gotNameIDN = %v, want %v", got.NameIDN, tt.wantNameIDN)
			}
			if got.NameUnicode != tt.wantNameUnicode {
				t.Errorf("MakeDomainFixForms() gotNameUnicode = %v, want %v", got.NameUnicode, tt.wantNameUnicode)
			}
			if got.UniqueName != tt.wantUniqueName {
				t.Errorf("MakeDomainFixForms() gotUniqueName = %v, want %v", got.UniqueName, tt.wantUniqueName)
			}
			if got.HasBang != tt.wantHasBang {
				t.Errorf("MakeDomainFixForms() gotHasTag = %v, want %v", got.HasBang, tt.wantHasBang)
			}
		})
	}
}
