package cfsingleredirect

import "testing"

func TestMakePageRuleBlob(t *testing.T) {
	type args struct {
		from     string
		to       string
		priority uint16
		code     uint16
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "one",
			args: args{"from", "to", 1, 301},
			want: "1,301,from,to",
		},
		{
			name: "two",
			args: args{"from", "to", 1, 9},
			want: "1,009,from,to",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakePageRuleBlob(tt.args.from, tt.args.to, tt.args.priority, tt.args.code); got != tt.want {
				t.Errorf("MakePageRuleBlob() = %v, want %v", got, tt.want)
			}
		})
	}
}
