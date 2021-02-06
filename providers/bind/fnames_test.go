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
	fmtBk1 := "db_%U"          // Something I've seen in books on DNS
	fmtBk2 := "db_%T_%D"       // Something I've seen in books on DNS
	fmtFancy := "%T%?_%D.zone" // Include the tag_ only if there is a tag
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
		{"bk1", args{fmtBk1, uu, dd, tt}, "db_uni"},
		{"bk2", args{fmtBk2, uu, dd, tt}, "db_tagy_domy"},
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
		{"u", args{"%U.zone"}, `(.*)!.+\.zone|(.*)\.zone`, false},
		{"d", args{"%D.zone"}, `(.*)\.zone`, false},
		{"basic", args{"%U - %T - %D"}, `(.*)!.+ - .* - (.*)|(.*) -  - (.*)`, false},
		{"bk1", args{"db_%U"}, `db_(.*)!.+|db_(.*)`, false},
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

	// A list of filenames one might find in a directory.
	filelist := []string{
		"foo!one.zone",                // u
		"dom.tld!two.zone",            // u
		"foo.zone",                    // d
		"dom.tld.zone",                // d
		"foo!one - one - foo",         // basic
		"dom.tld!two - two - dom.tld", // basic
		"db_foo",                      // bk1
		"db_dom.tld",                  // bk1
		"db_dom.tld!tag",              // bk1
		"db_inside_foo",               // bk2
		"db_outside_dom.tld",          // bk2
		"db__example.com",             // bk2
		"dom.zone",                    // fan
		"example.com.zone",            // fan (no tag)
		"mytag_example.com.zone",      // fan (w/ tag)
		"dom.zone",                    // fan (no tag)
		"mytag_dom.zone",              // fan (w/ tag)
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"0", args{"%D.zone", []string{"foo.zone", "dom.tld.zone"}}, []string{"foo", "dom.tld"}},
		{"1", args{"%U.zone", []string{"foo.zone", "dom.tld.zone"}}, []string{"foo", "dom.tld"}},
		{"2", args{"%T%?_%D.zone", []string{"inside_ex.tld.zone", "foo.zone", "dom.tld.zone"}}, []string{"ex.tld", "foo", "dom.tld"}},
		{"d", args{"%D.zone", filelist}, []string{"foo!one", "dom.tld!two", "foo", "dom.tld", "dom", "example.com", "mytag_example.com", "dom", "mytag_dom"}},
		{"u", args{"%U.zone", filelist}, []string{"foo", "dom.tld", "foo", "dom.tld", "dom", "example.com", "mytag_example.com", "dom", "mytag_dom"}},
		{"bk1", args{"db_%U", filelist}, []string{"foo", "dom.tld", "dom.tld", "inside_foo", "outside_dom.tld", "_example.com"}},
		{"bk2", args{"db_%T_%D", filelist}, []string{"foo", "dom.tld", "example.com"}},
		{"fan", args{"%T%?_%D.zone", filelist}, []string{"foo!one", "dom.tld!two", "foo", "dom.tld", "dom", "example.com", "example.com", "dom", "dom"}},
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
