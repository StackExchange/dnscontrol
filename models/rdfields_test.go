package models_test

import (
	"slices"
	"testing"

	dnsv2 "codeberg.org/miekg/dns"
	dnsrdatav2 "codeberg.org/miekg/dns/rdata"
	"github.com/DNSControl/dnscontrol/v5/models"
)

func TestRDtoFieldsJS(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s       dnsv2.RDATA
		want    []string
		wantErr bool
	}{
		{"mx", dnsrdatav2.MX{Preference: 10, Mx: "foo.example.com."}, []string{"10", `"foo.example.com."`}, false},
		{"csync one type", dnsrdatav2.CSYNC{Serial: 0, Flags: 1, TypeBitMap: []uint16{dnsv2.TypeNS}}, []string{"0", "1", `"NS"`}, false},
		{"csync many types", dnsrdatav2.CSYNC{Serial: 66, Flags: 3, TypeBitMap: []uint16{dnsv2.TypeA, dnsv2.TypeNS, dnsv2.TypeAAAA}}, []string{"66", "3", `"A NS AAAA"`}, false},
		{"csync unknown type", dnsrdatav2.CSYNC{Serial: 1, Flags: 0, TypeBitMap: []uint16{65534}}, []string{"1", "0", `"TYPE65534"`}, false},
		{"other []uint16 is not a type bitmap", uint16ListRD{Ports: []uint16{1, 2}}, []string{"[1 2]"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := models.RDtoFieldsJS(tt.s)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("RDtoFieldsJS() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("RDtoFieldsJS() succeeded unexpectedly")
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("RDtoFieldsJS() = %v, want %v", got, tt.want)
			}
		})
	}
}

// uint16ListRD is an RDATA whose []uint16 field is not a type bitmap.
type uint16ListRD struct {
	Ports []uint16
}

func (uint16ListRD) Len() int       { return 0 }
func (uint16ListRD) String() string { return "" }

func TestRDtoFieldsStrings(t *testing.T) {
	tests := []struct {
		name string
		s    dnsv2.RDATA
		want []string
	}{
		{"csync", dnsrdatav2.CSYNC{Serial: 66, Flags: 3, TypeBitMap: []uint16{dnsv2.TypeA, dnsv2.TypeNS, dnsv2.TypeAAAA}}, []string{"66", "3", "A NS AAAA"}},
		{"other []uint16 is not a type bitmap", uint16ListRD{Ports: []uint16{1, 2}}, []string{"[1 2]"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := models.RDtoFieldsStrings(tt.s)
			if err != nil {
				t.Fatalf("RDtoFieldsStrings() failed: %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("RDtoFieldsStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}
