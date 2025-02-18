package commands

import (
	"testing"

	"github.com/StackExchange/dnscontrol/v4/models"
)

func Test_whichZonesToProcess(t *testing.T) {

	dcNoTag := &models.DomainConfig{Name: "example.com"}
	dcTaggedGeorge := &models.DomainConfig{Name: "example.com!george"}
	dcTaggedJohn := &models.DomainConfig{Name: "example.com!john"}

	allDC := []*models.DomainConfig{dcNoTag, dcTaggedGeorge, dcTaggedJohn}
	for _, dc := range allDC {
		dc.UpdateSplitHorizonNames()
	}

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
			name: "testMultiFilterTaggedNoMatch",
			why:  "Should return nothing",
			args: args{
				dc:     allDC,
				filter: "example.com!ringo",
			},
			want: []*models.DomainConfig{},
		},
		{
			name: "testMultiFilterTaggedWildcard",
			why:  "Should return two tagged domains",
			args: args{
				dc:     allDC,
				filter: "example.com!*",
			},
			want: []*models.DomainConfig{dcTaggedGeorge, dcTaggedJohn},
		},
		{
			name: "testFilterEmptyTag",
			why:  "Should return untagged domain",
			args: args{
				dc:     allDC,
				filter: "example.com!",
			},
			want: []*models.DomainConfig{dcNoTag},
		},
		{
			name: "testFilterNoTagNoMatch",
			why:  "Should return nothing",
			args: args{
				dc:     []*models.DomainConfig{dcTaggedGeorge, dcTaggedJohn},
				filter: "example.com",
			},
			want: []*models.DomainConfig{},
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
			name: "testFilterNoTagTagged",
			why:  "Should return the tagged and untagged domains",
			args: args{
				dc:     allDC,
				filter: "example.com!george,example.com",
			},
			want: []*models.DomainConfig{dcTaggedGeorge, dcNoTag},
		},
		{
			name: "testAllFilter",
			why:  "Should return all domain configs",
			args: args{
				dc:     allDC,
				filter: "all",
			},
			want: allDC,
		},
		{
			name: "testNoFilter",
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
			if len(got) != len(tt.want) {
				t.Errorf("whichZonesToProcess(): %s", tt.why)
				return
			}
			for i := range got {
				if got[i].Name != tt.want[i].Name {
					t.Errorf("whichZonesToProcess(): %s", tt.why)
				}
			}
		})
	}
}
