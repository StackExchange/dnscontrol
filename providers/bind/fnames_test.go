package bind

import (
	"reflect"
	"testing"
)

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
		{"errP", args{fmtErrorPct, uu, dd, tt}, "literal%(format may not end in %)"},
		{"errQ", args{fmtErrorOpt, uu, dd, tt}, "literal%(format may not end in %?)"},
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

func Test_makeExtractor(t *testing.T) {
	type args struct {
		format string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"u", args{"%U.zone"}, `(.*)!?\.zone`, false},
		{"d", args{"%D.zone"}, `(.*)\.zone`, false},
		{"basic", args{"%U - %T - %D"}, `(.*)!? - .* - (.*)`, false},
		{"bk1", args{"db_%U"}, `db_(.*)!?`, false},
		{"bk2", args{"db_%T_%D"}, `db_.*_(.*)`, false},
		{"fan", args{"%T%?_%D.zone"}, `.*_(.*)\.zone|(.*)\.zone`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeExtractor(tt.args.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeExtractor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("makeExtractor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractZonesFromFilenames(t *testing.T) {
	type args struct {
		format string
		names  []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"0", args{"%D.zone", []string{"foo.zone", "dom.tld.zone"}}, []string{"foo", "dom.tld"}},
		{"1", args{"%U.zone", []string{"foo.zone", "dom.tld.zone"}}, []string{"foo", "dom.tld"}},
		{"2", args{"%T%?_%D.zone", []string{
			"inside_ex.tld.zone",
			"foo.zone",
			"dom.tld.zone",
		}}, []string{
			"ex.tld",
			"foo",
			"dom.tld",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractZonesFromFilenames(tt.args.format, tt.args.names); !reflect.DeepEqual(got, tt.want) {
				ext, _ := makeExtractor(tt.args.format)
				t.Errorf("extractZonesFromFilenames() = %v, want %v Fm=%s Ex=%s", got, tt.want, tt.args.format, ext)
			}
		})
	}
}
