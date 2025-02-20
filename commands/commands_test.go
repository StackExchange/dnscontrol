package commands

import "testing"

func Test_domainInList(t *testing.T) {
	type args struct {
		domain string
		list   []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "small",
			args: args{
				domain: "foo.com",
				list:   []string{"foo.com"},
			},
			want: true,
		},
		{
			name: "big",
			args: args{
				domain: "foo.com",
				list:   []string{"example.com", "foo.com", "baz.com"},
			},
			want: true,
		},
		{
			name: "missing",
			args: args{
				domain: "foo.com",
				list:   []string{"bar.com"},
			},
			want: false,
		},
		{
			name: "wildcard",
			args: args{
				domain: "*.10.in-addr.arpa",
				list:   []string{"bar.com", "10.in-addr.arpa", "example.com"},
			},
			want: false,
		},
		{
			name: "wildcardmissing",
			args: args{
				domain: "*.10.in-addr.arpa",
				list:   []string{"bar.com", "6.in-addr.arpa", "example.com"},
			},
			want: false,
		},
		{
			name: "tagged",
			args: args{
				domain: "foo.com!bar",
				list:   []string{"foo.com"},
			},
			want: false,
		},
		{
			name: "taggedWildcard",
			args: args{
				domain: "foo.com!bar",
				list:   []string{"foo.com!*"},
			},
			want: true,
		},
		{
			name: "taggedWildcardMatchesEmpty",
			args: args{
				domain: "foo.com!",
				list:   []string{"foo.com!*"},
			},
			want: true,
		},
		{
			name: "taggedWildcardNotMatchUntagged",
			args: args{
				domain: "foo.com",
				list:   []string{"foo.com!*"},
			},
			want: false,
		},
		{
			name: "taggedEmtpy",
			args: args{
				domain: "foo.com",
				list:   []string{"foo.com!"},
			},
			want: true,
		},
		{
			name: "domainTaggedEmtpy",
			args: args{
				domain: "foo.com!",
				list:   []string{"foo.com"},
			},
			want: true,
		},
		{
			name: "filterTaggedNoMatch",
			args: args{
				domain: "foo.com",
				list:   []string{"foo.com!foo"},
			},
			want: false,
		},
		{
			name: "domainTaggedNoMatch",
			args: args{
				domain: "foo.com!foo",
				list:   []string{"foo.com"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domainInList(tt.args.domain, tt.args.list); got != tt.want {
				t.Errorf("domainInList() = %v, want %v", got, tt.want)
			}
		})
	}
}
