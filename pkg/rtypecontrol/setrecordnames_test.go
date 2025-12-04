package rtypecontrol

import (
	"testing"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/domaintags"
)

func TestSetRecordNames(t *testing.T) {
	dc := &domaintags.DomainNameVarieties{
		NameASCII:   "example.com",
		NameRaw:     "example.com",
		NameUnicode: "example.com",
	}
	dcIDN := &domaintags.DomainNameVarieties{
		NameASCII:   "xn--bcher-kva.com",
		NameRaw:     "bücher.com",
		NameUnicode: "bücher.com",
	}

	tests := []struct {
		name        string
		rec         *models.RecordConfig
		dc          *domaintags.DomainNameVarieties
		n           string
		expectedRec *models.RecordConfig
	}{
		{
			name: "normal_at",
			rec:  &models.RecordConfig{},
			dc:   dc,
			n:    "@",
			expectedRec: &models.RecordConfig{
				Name:            "@",
				NameRaw:         "@",
				NameUnicode:     "@",
				NameFQDN:        "example.com",
				NameFQDNRaw:     "example.com",
				NameFQDNUnicode: "example.com",
			},
		},
		{
			name: "normal_label",
			rec:  &models.RecordConfig{},
			dc:   dc,
			n:    "www",
			expectedRec: &models.RecordConfig{
				Name:            "www",
				NameRaw:         "www",
				NameUnicode:     "www",
				NameFQDN:        "www.example.com",
				NameFQDNRaw:     "www.example.com",
				NameFQDNUnicode: "www.example.com",
			},
		},
		{
			name: "normal_idn_label",
			rec:  &models.RecordConfig{},
			dc:   dc,
			n:    "bücher",
			expectedRec: &models.RecordConfig{
				Name:            "xn--bcher-kva",
				NameRaw:         "bücher",
				NameUnicode:     "bücher",
				NameFQDN:        "xn--bcher-kva.example.com",
				NameFQDNRaw:     "bücher.example.com",
				NameFQDNUnicode: "bücher.example.com",
			},
		},
		{
			name: "normal_idn_domain",
			rec:  &models.RecordConfig{},
			dc:   dcIDN,
			n:    "www",
			expectedRec: &models.RecordConfig{
				Name:            "www",
				NameRaw:         "www",
				NameUnicode:     "www",
				NameFQDN:        "www.xn--bcher-kva.com",
				NameFQDNRaw:     "www.bücher.com",
				NameFQDNUnicode: "www.bücher.com",
			},
		},
		{
			name: "extend_at",
			rec:  &models.RecordConfig{SubDomain: "sub"},
			dc:   dc,
			n:    "@",
			expectedRec: &models.RecordConfig{
				SubDomain:       "sub",
				Name:            "sub",
				NameRaw:         "sub",
				NameUnicode:     "sub",
				NameFQDN:        "sub.example.com",
				NameFQDNRaw:     "sub.example.com",
				NameFQDNUnicode: "sub.example.com",
			},
		},
		{
			name: "extend_label",
			rec:  &models.RecordConfig{SubDomain: "sub"},
			dc:   dc,
			n:    "www",
			expectedRec: &models.RecordConfig{
				SubDomain:       "sub",
				Name:            "www.sub",
				NameRaw:         "www.sub",
				NameUnicode:     "www.sub",
				NameFQDN:        "www.sub.example.com",
				NameFQDNRaw:     "www.sub.example.com",
				NameFQDNUnicode: "www.sub.example.com",
			},
		},
		{
			name: "extend_idn_subdomain",
			rec:  &models.RecordConfig{SubDomain: "bücher"},
			dc:   dc,
			n:    "www",
			expectedRec: &models.RecordConfig{
				SubDomain:       "bücher",
				Name:            "www.xn--bcher-kva",
				NameRaw:         "www.bücher",
				NameUnicode:     "www.bücher",
				NameFQDN:        "www.xn--bcher-kva.example.com",
				NameFQDNRaw:     "www.bücher.example.com",
				NameFQDNUnicode: "www.bücher.example.com",
			},
		},
		{
			name: "extend_idn_label",
			rec:  &models.RecordConfig{SubDomain: "sub"},
			dc:   dc,
			n:    "bücher",
			expectedRec: &models.RecordConfig{
				SubDomain:       "sub",
				Name:            "xn--bcher-kva.sub",
				NameRaw:         "bücher.sub",
				NameUnicode:     "bücher.sub",
				NameFQDN:        "xn--bcher-kva.sub.example.com",
				NameFQDNRaw:     "bücher.sub.example.com",
				NameFQDNUnicode: "bücher.sub.example.com",
			},
		},
		{
			name: "extend_idn_subdomain_and_label",
			rec:  &models.RecordConfig{SubDomain: "bücher"},
			dc:   dc,
			n:    "könig",
			expectedRec: &models.RecordConfig{
				SubDomain:       "bücher",
				Name:            "xn--knig-5qa.xn--bcher-kva",
				NameRaw:         "könig.bücher",
				NameUnicode:     "könig.bücher",
				NameFQDN:        "xn--knig-5qa.xn--bcher-kva.example.com",
				NameFQDNRaw:     "könig.bücher.example.com",
				NameFQDNUnicode: "könig.bücher.example.com",
			},
		},
		{
			name: "extend_idn_domain_and_subdomain",
			rec:  &models.RecordConfig{SubDomain: "bücher"},
			dc:   dcIDN,
			n:    "www",
			expectedRec: &models.RecordConfig{
				SubDomain:       "bücher",
				Name:            "www.xn--bcher-kva",
				NameRaw:         "www.bücher",
				NameUnicode:     "www.bücher",
				NameFQDN:        "www.xn--bcher-kva.xn--bcher-kva.com",
				NameFQDNRaw:     "www.bücher.bücher.com",
				NameFQDNUnicode: "www.bücher.bücher.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setRecordNames(tt.rec, tt.dc, tt.n)
			if tt.rec.Name != tt.expectedRec.Name {
				t.Errorf("Name: got %q, want %q", tt.rec.Name, tt.expectedRec.Name)
			}
			if tt.rec.NameRaw != tt.expectedRec.NameRaw {
				t.Errorf("NameRaw: got %q, want %q", tt.rec.NameRaw, tt.expectedRec.NameRaw)
			}
			if tt.rec.NameUnicode != tt.expectedRec.NameUnicode {
				t.Errorf("NameUnicode: got %q, want %q", tt.rec.NameUnicode, tt.expectedRec.NameUnicode)
			}
			if tt.rec.NameFQDN != tt.expectedRec.NameFQDN {
				t.Errorf("NameFQDN: got %q, want %q", tt.rec.NameFQDN, tt.expectedRec.NameFQDN)
			}
			if tt.rec.NameFQDNRaw != tt.expectedRec.NameFQDNRaw {
				t.Errorf("NameFQDNRaw: got %q, want %q", tt.rec.NameFQDNRaw, tt.expectedRec.NameFQDNRaw)
			}
			if tt.rec.NameFQDNUnicode != tt.expectedRec.NameFQDNUnicode {
				t.Errorf("NameFQDNUnicode: got %q, want %q", tt.rec.NameFQDNUnicode, tt.expectedRec.NameFQDNUnicode)
			}
		})
	}
}
