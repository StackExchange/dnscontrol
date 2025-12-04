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
		wantNameASCII   string
		wantNameUnicode string
		wantUniqueName  string
		wantHasBang     bool
	}{
		{
			name:            "simple domain",
			input:           "example.com",
			wantTag:         "",
			wantNameRaw:     "example.com",
			wantNameASCII:   "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com",
			wantHasBang:     false,
		},
		{
			name:            "domain with tag",
			input:           "example.com!mytag",
			wantTag:         "mytag",
			wantNameRaw:     "example.com",
			wantNameASCII:   "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com!mytag",
			wantHasBang:     true,
		},
		{
			name:            "domain with empty tag",
			input:           "example.com!",
			wantTag:         "",
			wantNameRaw:     "example.com",
			wantNameASCII:   "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com!",
			wantHasBang:     true,
		},
		{
			name:            "unicode domain",
			input:           "उदाहरण.com",
			wantTag:         "",
			wantNameRaw:     "उदाहरण.com",
			wantNameASCII:   "xn--p1b6ci4b4b3a.com",
			wantNameUnicode: "उदाहरण.com",
			wantUniqueName:  "xn--p1b6ci4b4b3a.com",
			wantHasBang:     false,
		},
		{
			name:            "unicode domain with tag",
			input:           "उदाहरण.com!mytag",
			wantTag:         "mytag",
			wantNameRaw:     "उदाहरण.com",
			wantNameASCII:   "xn--p1b6ci4b4b3a.com",
			wantNameUnicode: "उदाहरण.com",
			wantUniqueName:  "xn--p1b6ci4b4b3a.com!mytag",
			wantHasBang:     true,
		},
		{
			name:            "punycode domain",
			input:           "xn--p1b6ci4b4b3a.com",
			wantTag:         "",
			wantNameRaw:     "xn--p1b6ci4b4b3a.com",
			wantNameASCII:   "xn--p1b6ci4b4b3a.com",
			wantNameUnicode: "उदाहरण.com",
			wantUniqueName:  "xn--p1b6ci4b4b3a.com",
			wantHasBang:     false,
		},
		{
			// Unicode chars should be left alone (as far as case folding goes)
			// Here are some Armenian characters https://tools.lgm.cl/lettercase.html
			name:            "mixed case unicode",
			input:           "fooԷէԸըԹ.com!myTag",
			wantTag:         "myTag",
			wantNameRaw:     "fooԷէԸըԹ.com",
			wantNameASCII:   "xn--foo-b7dfg43aja.com",
			wantNameUnicode: "fooԷէԸըԹ.com",
			wantUniqueName:  "xn--foo-b7dfg43aja.com!myTag",
			wantHasBang:     true,
		},
		{
			name:            "punycode domain with tag",
			input:           "xn--p1b6ci4b4b3a.com!mytag",
			wantTag:         "mytag",
			wantNameRaw:     "xn--p1b6ci4b4b3a.com",
			wantNameASCII:   "xn--p1b6ci4b4b3a.com",
			wantNameUnicode: "उदाहरण.com",
			wantUniqueName:  "xn--p1b6ci4b4b3a.com!mytag",
			wantHasBang:     true,
		},
		{
			name:            "mixed case domain",
			input:           "Example.COM",
			wantTag:         "",
			wantNameRaw:     "Example.COM",
			wantNameASCII:   "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com",
			wantHasBang:     false,
		},
		{
			name:            "mixed case domain with tag",
			input:           "Example.COM!MyTag",
			wantTag:         "MyTag",
			wantNameRaw:     "Example.COM",
			wantNameASCII:   "example.com",
			wantNameUnicode: "example.com",
			wantUniqueName:  "example.com!MyTag",
			wantHasBang:     true,
		},
		// This is used in the documentation for the BIND provider, thus we test
		// it to make sure we got it right.
		{
			name:            "BIND example 1",
			input:           "рф.com!myTag",
			wantTag:         "myTag",
			wantNameRaw:     "рф.com",
			wantNameASCII:   "xn--p1ai.com",
			wantNameUnicode: "рф.com",
			wantUniqueName:  "xn--p1ai.com!myTag",
			wantHasBang:     true,
		},
		{
			name:            "BIND example 2",
			input:           "рф.com",
			wantTag:         "",
			wantNameRaw:     "рф.com",
			wantNameASCII:   "xn--p1ai.com",
			wantNameUnicode: "рф.com",
			wantUniqueName:  "xn--p1ai.com",
			wantHasBang:     false,
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
			if got.NameASCII != tt.wantNameASCII {
				t.Errorf("MakeDomainFixForms() gotNameASCII = %v, want %v", got.NameASCII, tt.wantNameASCII)
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
