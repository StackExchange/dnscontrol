package credsfile

import (
	"reflect"
	"testing"
)

func Test_keysWithColons(t *testing.T) {
	type args struct {
		list []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"0", args{list: []string{""}}, nil},
		{"1", args{list: []string{"none"}}, nil},
		{"2", args{list: []string{"a:b"}}, []string{"a:b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keysWithColons(tt.args.list); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("keysWithColons() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_quotedList(t *testing.T) {
	type args struct {
		l []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"none", args{}, ""},
		{"single", args{l: []string{"one"}}, `"one"`},
		{"two", args{l: []string{"ricky", "lucy"}}, `"ricky", "lucy"`},
		{"three", args{l: []string{"manny", "moe", "jack"}}, `"manny", "moe", "jack"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := quotedList(tt.args.l); got != tt.want {
				t.Errorf("quotedList() = %v, want %v", got, tt.want)
			}
		})
	}
}
