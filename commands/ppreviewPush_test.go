package commands

import (
	"testing"

	"github.com/StackExchange/dnscontrol/v4/models"
)

// makeDomainConfig takes a domain name and returns a DomainConfig
func makeDomainConfig(domainName string) *models.DomainConfig {
	dc := &models.DomainConfig{
		Name: domainName,
	}
	dc.UpdateSplitHorizonNames()
	return dc
}

// wantEqualDC compares two slices of DomainConfig and returns true if their names match
func wantEqualDC(a, b []*models.DomainConfig) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name {
			return false
		}
	}
	return true
}

func Test_whichZonesToProcess(t *testing.T) {

	dcNoTag := makeDomainConfig("example.com")
	dcTaggedGeorge := makeDomainConfig("example.com!george")
	dcTaggedJohn := makeDomainConfig("example.com!john")
	allDC := []*models.DomainConfig{dcNoTag, dcTaggedGeorge, dcTaggedJohn}

	type args struct {
		dc     []*models.DomainConfig
		filter string
	}

	tests := []struct {
		name string
		why  string
		args args
		want []*models.DomainConfig
	}{
		{
			name: "testFilterTagged",
			why:  "Should return one tagged domain",
			args: args{
				dc:     allDC,
				filter: "example.com!george",
			},
			want: []*models.DomainConfig{dcTaggedGeorge},
		},
		{
			name: "testMultiFilterTagged",
			why:  "Should return two tagged domains",
			args: args{
				dc:     allDC,
				filter: "example.com!george,example.com!john",
			},
			want: []*models.DomainConfig{dcTaggedGeorge, dcTaggedJohn},
		},
		{
			name: "testFilterNoTag",
			why:  "Should return the non-tagged domain",
			args: args{
				dc:     allDC,
				filter: "example.com",
			},
			want: []*models.DomainConfig{dcNoTag},
		},
		{
			name: "testFilterTaggedNoMatch",
			why:  "Should return nothing",
			args: args{
				dc:     []*models.DomainConfig{dcNoTag},
				filter: "example.com!george",
			},
			want: []*models.DomainConfig{},
		},
		{
			name: "testBothFilterNoTagTagged",
			why:  "Should return the tagged and untagged domains",
			args: args{
				dc:     allDC,
				filter: "example.com!george,example.com",
			},
			want: []*models.DomainConfig{dcTaggedGeorge, dcNoTag},
		},
		{
			name: "testBothFilterNoTag",
			why:  "Should return the non-tagged domain",
			args: args{
				dc:     allDC,
				filter: "example.com",
			},
			want: []*models.DomainConfig{dcNoTag},
		},
		{
			name: "testBothAllFilter",
			why:  "Should return all domain configs",
			args: args{
				dc:     allDC,
				filter: "all",
			},
			want: allDC,
		},
		{
			name: "testBothNoFilter",
			why:  "Should return all domain configs",
			args: args{
				dc:     allDC,
				filter: "",
			},
			want: allDC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := whichZonesToProcess(tt.args.dc, tt.args.filter)
			if !wantEqualDC(got, tt.want) {
				t.Errorf("whichZonesToProcess(): %s", tt.why)
			}
		})
	}
}
