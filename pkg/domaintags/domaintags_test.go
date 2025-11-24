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
	}{
		{
			name:            "simple domain",
			input:           "example.com",
			wantTag:         "",
			wantNameRaw:     "example.com",
			wantNameIDN:     "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com!",
		},
		{
			name:            "domain with tag",
			input:           "example.com!mytag",
			wantTag:         "mytag",
			wantNameRaw:     "example.com",
			wantNameIDN:     "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com!mytag",
		},
		{
			name:            "domain with empty tag",
			input:           "example.com!",
			wantTag:         "",
			wantNameRaw:     "example.com",
			wantNameIDN:     "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com!",
		},
		{
			name:            "unicode domain",
			input:           "उदाहरण.com",
			wantTag:         "",
			wantNameRaw:     "उदाहरण.com",
			wantNameIDN:     "xn--p1b6ci4b4b3a.com",
			wantNameUnicode: "उदाहरण.com",
			wantUniqueName:  "xn--p1b6ci4b4b3a.com!",
		},
		{
			name:            "unicode domain with tag",
			input:           "उदाहरण.com!mytag",
			wantTag:         "mytag",
			wantNameRaw:     "उदाहरण.com",
			wantNameIDN:     "xn--p1b6ci4b4b3a.com",
			wantNameUnicode: "उदाहरण.com",
			wantUniqueName:  "xn--p1b6ci4b4b3a.com!mytag",
		},
		{
			name:            "punycode domain",
			input:           "xn--p1b6ci4b4b3a.com",
			wantTag:         "",
			wantNameRaw:     "xn--p1b6ci4b4b3a.com",
			wantNameIDN:     "xn--p1b6ci4b4b3a.com",
			wantNameUnicode: "उदाहरण.com",
			wantUniqueName:  "xn--p1b6ci4b4b3a.com!",
		},
		{
			name:            "punycode domain with tag",
			input:           "xn--p1b6ci4b4b3a.com!mytag",
			wantTag:         "mytag",
			wantNameRaw:     "xn--p1b6ci4b4b3a.com",
			wantNameIDN:     "xn--p1b6ci4b4b3a.com",
			wantNameUnicode: "उदाहरण.com",
			wantUniqueName:  "xn--p1b6ci4b4b3a.com!mytag",
		},
		{
			name:            "mixed case domain",
			input:           "Example.COM",
			wantTag:         "",
			wantNameRaw:     "example.com",
			wantNameIDN:     "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com!",
		},
		{
			name:            "mixed case domain with tag",
			input:           "Example.COM!MyTag",
			wantTag:         "MyTag",
			wantNameRaw:     "example.com",
			wantNameIDN:     "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com!MyTag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTag, gotNameRaw, gotNameIDN, gotNameUnicode, gotUniqueName := MakeDomainFixForms(tt.input)
			if gotTag != tt.wantTag {
				t.Errorf("MakeDomainFixForms() gotTag = %v, want %v", gotTag, tt.wantTag)
			}
			if gotNameRaw != tt.wantNameRaw {
				t.Errorf("MakeDomainFixForms() gotNameRaw = %v, want %v", gotNameRaw, tt.wantNameRaw)
			}
			if gotNameIDN != tt.wantNameIDN {
				t.Errorf("MakeDomainFixForms() gotNameIDN = %v, want %v", gotNameIDN, tt.wantNameIDN)
			}
			if gotNameUnicode != tt.wantNameUnicode {
				t.Errorf("MakeDomainFixForms() gotNameUnicode = %v, want %v", gotNameUnicode, tt.wantNameUnicode)
			}
			if gotUniqueName != tt.wantUniqueName {
				t.Errorf("MakeDomainFixForms() gotUniqueName = %v, want %v", gotUniqueName, tt.wantUniqueName)
			}
		})
	}
}
