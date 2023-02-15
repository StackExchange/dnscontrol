package models

// var rmapNil map[string]bool
// var rmapAMX = map[string]bool{
// 	"A":  true,
// 	"MX": true,
// }
// var rmapCNAME = map[string]bool{
// 	"CNAME": true,
// }

// func makeRec(label, rtype, content string) *RecordConfig {
// 	origin := "f.com"
// 	r := RecordConfig{TTL: 300}
// 	r.SetLabel(label, origin)
// 	r.PopulateFromString(rtype, content, origin)
// 	return &r
// }

// func Test_match(t *testing.T) {

// 	testRecLammaA1234 := makeRec("lamma", "A", "1.2.3.4")

// 	type args struct {
// 		rc       *RecordConfig
// 		glabel   glob.Glob
// 		gtarget  glob.Glob
// 		hasRType map[string]bool
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want bool
// 	}{

// 		{
// 			name: "match3",
// 			args: args{
// 				rc:       testRecLammaA1234,
// 				glabel:   glob.MustCompile("lam*"),
// 				hasRType: rmapAMX,
// 				gtarget:  glob.MustCompile("1.2.3.*"),
// 			},
// 			want: true,
// 		},

// 		{
// 			name: "match2",
// 			args: args{
// 				rc:       testRecLammaA1234,
// 				glabel:   glob.MustCompile("lam*"),
// 				hasRType: rmapAMX,
// 				gtarget:  nil,
// 			},
// 			want: true,
// 		},

// 		{
// 			name: "match1",
// 			args: args{
// 				rc:       testRecLammaA1234,
// 				glabel:   glob.MustCompile("lam*"),
// 				hasRType: rmapNil,
// 				gtarget:  nil,
// 			},
// 			want: true,
// 		},

// 		{
// 			name: "reject1",
// 			args: args{
// 				rc:       testRecLammaA1234,
// 				glabel:   glob.MustCompile("yyyy"),
// 				hasRType: rmapAMX,
// 				gtarget:  glob.MustCompile("1.2.3.*"),
// 			},
// 			want: false,
// 		},

// 		{
// 			name: "reject2",
// 			args: args{
// 				rc:       testRecLammaA1234,
// 				glabel:   glob.MustCompile("lam*"),
// 				hasRType: rmapCNAME,
// 				gtarget:  glob.MustCompile("1.2.3.*"),
// 			},
// 			want: false,
// 		},

// 		{
// 			name: "reject3",
// 			args: args{
// 				rc:       testRecLammaA1234,
// 				glabel:   glob.MustCompile("lam*"),
// 				hasRType: rmapAMX,
// 				gtarget:  glob.MustCompile("zzzzz"),
// 			},
// 			want: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := match(tt.args.rc, tt.args.glabel, tt.args.gtarget, tt.args.hasRType); got != tt.want {
// 				t.Errorf("match() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_matchType(t *testing.T) {
// 	type args struct {
// 		s        string
// 		hasRType map[string]bool
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want bool
// 	}{

// 		{
// 			name: "matchCNAME",
// 			args: args{"CNAME", rmapCNAME},
// 			want: true,
// 		},

// 		{
// 			name: "rejectCNAME",
// 			args: args{"MX", rmapCNAME},
// 			want: false,
// 		},

// 		{
// 			name: "matchNIL",
// 			args: args{"CNAME", rmapNil},
// 			want: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := matchType(tt.args.s, tt.args.hasRType); got != tt.want {
// 				t.Errorf("matchType() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
