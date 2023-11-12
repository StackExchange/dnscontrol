package prettyzone

import "testing"

func Test_txtToNative(t *testing.T) {
	type args struct {
		parts []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1", args{[]string{`foo`}}, `"foo"`},
		{"2", args{[]string{`one`, `two`}}, `"one" "two"`},
		{"single", args{[]string{`sin'gle`}}, `"sin'gle"`},
		{"double", args{[]string{`dou"ble`}}, `"dou\"ble"`},
		{"outer", args{[]string{`"outer"`}}, `"\"outer\""`},
		{"backtick", args{[]string{"back`tick"}}, "\"back`tick\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := txtToNative(tt.args.parts); got != tt.want {
				t.Errorf("txtToNative() = %v, want %v", got, tt.want)
			}
		})
	}
}
