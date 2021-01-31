package bind

import "testing"

func Test_makeFileName(t *testing.T) {

	uu := "uni"
	dd := "domy"
	tt := "tagy"
	fmtDefault := "%U.zone"
	fmtBasic := "%U - %T - %D"
	fmtBook1 := "db_%U"
	fmtBook2 := "db_%T_%D"
	fmtFancy := "%T%?_%D.zone"
	fmtErrorPct := "literal%"
	fmtErrorOpt := "literal%?"
	fmtErrorUnk := "literal%o" // Unknown % verb

	type args struct {
		format     string
		uniquename string
		domain     string
		tag        string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"literal", args{"literal", uu, dd, tt}, "literal"},
		{"basic", args{fmtBasic, uu, dd, tt}, "uni - tagy - domy"},
		{"solo", args{"%D", uu, dd, tt}, "domy"},
		{"front", args{"%Daaa", uu, dd, tt}, "domyaaa"},
		{"tail", args{"bbb%D", uu, dd, tt}, "bbbdomy"},
		{"def", args{fmtDefault, uu, dd, tt}, "uni.zone"},
		{"bk1", args{fmtBook1, uu, dd, tt}, "db_uni"},
		{"bk2", args{fmtBook2, uu, dd, tt}, "db_tagy_domy"},
		{"fanWI", args{fmtFancy, uu, dd, tt}, "tagy_domy.zone"},
		{"fanWO", args{fmtFancy, uu, dd, ""}, "domy.zone"},
		{"errP", args{fmtErrorPct, uu, dd, tt}, "literal%(string may not end in %)"},
		{"errQ", args{fmtErrorOpt, uu, dd, tt}, "literal%(string may not end in %?)"},
		{"errU", args{fmtErrorUnk, uu, dd, tt}, "literal%(unknown %verb %o)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeFileName(tt.args.format, tt.args.uniquename, tt.args.domain, tt.args.tag); got != tt.want {
				t.Errorf("makeFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}
