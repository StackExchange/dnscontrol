package diff2

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/stretchr/testify/assert"
)

var e0 = makeRec("laba", "A", "1.2.3.4")       //      [0]
var e1 = makeRec("laba", "MX", "10 laba")      //     [1]
var e2 = makeRec("labc", "CNAME", "laba")      //     [2]
var e3 = makeRec("labe", "A", "10.10.10.15")   //  [3]
var e4 = makeRec("labe", "A", "10.10.10.16")   //  [4]
var e5 = makeRec("labe", "A", "10.10.10.17")   //  [5]
var e6 = makeRec("labe", "A", "10.10.10.18")   //  [6]
var e7 = makeRec("labg", "NS", "10.10.10.15")  // [7]
var e8 = makeRec("labg", "NS", "10.10.10.16")  // [8]
var e9 = makeRec("labg", "NS", "10.10.10.17")  // [9]
var e10 = makeRec("labg", "NS", "10.10.10.18") // [10]
var e11 = makeRec("labh", "CNAME", "labd")     //     [11]

var d0 = makeRec("laba", "A", "1.2.3.4")       //      [0']
var d1 = makeRec("laba", "A", "1.2.3.5")       //      [1']
var d2 = makeRec("laba", "MX", "20 labb")      //     [2']
var d3 = makeRec("labe", "A", "10.10.10.95")   //  [3']
var d4 = makeRec("labe", "A", "10.10.10.96")   //  [4']
var d5 = makeRec("labe", "A", "10.10.10.97")   //  [5']
var d6 = makeRec("labe", "A", "10.10.10.98")   //  [6']
var d7 = makeRec("labf", "TXT", "foo")         //      [7']
var d8 = makeRec("labg", "NS", "10.10.10.10")  // [8']
var d9 = makeRec("labg", "NS", "10.10.10.15")  // [9']
var d10 = makeRec("labg", "NS", "10.10.10.16") // [10']
var d11 = makeRec("labg", "NS", "10.10.10.97") // [11']
var d12 = makeRec("labh", "A", "1.2.3.4")      //      [12']

var e0tc = targetConfig{compareable: "1.2.3.4 ttl=0", rec: e0}
var d0tc = targetConfig{compareable: "1.2.3.4 ttl=0", rec: d0}
var d1tc = targetConfig{compareable: "1.2.3.5 ttl=0", rec: d1}

func makeChange(v Verb, l, t string, old, new models.Records, msgs []string) Change {
	c := Change{
		Type: v,
		Old:  old,
		New:  new,
		Msgs: msgs,
	}
	c.Key.NameFQDN = l
	c.Key.Type = t
	return c
}

// func xTest_analyzeByRecordSet(t *testing.T) {
// 	origin := "f.com"
// 	existing := models.Records{e0, e1, e2, e3, e4, e5, e6, e7, e8, e9, e10, e11}
// 	desired := models.Records{d0, d1, d2, d3, d4, d5, d6, d7, d8, d9, d10, d11, d12}
// 	cc := NewCompareConfig(origin, existing, desired, nil)
// 	want := makeChange(CHANGE, "laba.f.com", "A", nil, models.Records{e0, d1}, []string{"CHANGE"})
// 	t.Run("01", func(t *testing.T) {
// 		if got := analyzeByRecordSet(cc); !reflect.DeepEqual(got, want) {
// 			t.Errorf("analyzeByRecordSet()\ngot =%+v\nwant=%+v", got, want)
// 		}
// 	})
// }

func Test_analyzeByRecordSet(t *testing.T) {
	type args struct {
		cc *CompareConfig
	}

	origin := "f.com"
	//existing := models.Records{e0, e1, e2, e3, e4, e5, e6, e7, e8, e9, e10, e11}
	//desired := models.Records{d0, d1, d2, d3, d4, d5, d6, d7, d8, d9, d10, d11, d12}
	//	cc0 := NewCompareConfig(origin, models.Records{e0}, models.Records{d0}, nil)
	//	cc1 := NewCompareConfig(origin, existing, desired, nil)

	tests := []struct {
		name string
		args args
		want []string
	}{

		{
			name: "oneequal",
			args: args{
				cc: NewCompareConfig(origin,
					models.Records{e0},
					models.Records{d0},
					nil),
			},
			want: []string{},
		},

		// {
		// 	name: "sameequal",
		// 	args: args{
		// 		cc: NewCompareConfig(origin,
		// 			models.Records{e0, e1},
		// 			models.Records{d0, d2},
		// 			nil),
		// 	},
		// 	want: []string{"CHANGE laba.f.com MX 20 labb"},
		// },

		// {
		// 	name: "big",
		// 	args: args{cc1},
		// 	want: []string{
		// 		"CHANGE laba.f.com MX 20 labb",
		// 		"DELETE labc.f.com CNAME laba",
		// 		"DELETE labe.f.com A 10.10.10.15",
		// 		"DELETE labe.f.com A 10.10.10.16",
		// 		"DELETE labe.f.com A 10.10.10.17",
		// 		"DELETE labe.f.com A 10.10.10.18",
		// 		"DELETE labg.f.com NS 10.10.10.15",
		// 		"DELETE labg.f.com NS 10.10.10.16",
		// 		"DELETE labg.f.com NS 10.10.10.17",
		// 		"DELETE labg.f.com NS 10.10.10.18",
		// 		"CHANGE laba.f.com A 1.2.3.4",
		// 		"CREATE laba.f.com A 1.2.3.5",
		// 		"CREATE labe.f.com A 10.10.10.95",
		// 		"CREATE labe.f.com A 10.10.10.96",
		// 		"CREATE labe.f.com A 10.10.10.97",
		// 		"CREATE labe.f.com A 10.10.10.98",
		// 		"CREATE laba.f.com MX 20 labb",
		// 		"CREATE labf.f.com TXT \"foo\"",
		// 		"CREATE labg.f.com NS 10.10.10.10",
		// 		"CREATE labg.f.com NS 10.10.10.15",
		// 		"CREATE labg.f.com NS 10.10.10.16",
		// 		"CREATE labg.f.com NS 10.10.10.97",
		// 		"CREATE labh.f.com A 1.2.3.4",
		// 	},
		// },

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := justMsgs(analyzeByRecordSet(tt.args.cc))
			//fmt.Printf("ABRS %02d: got=%q\n", i, got)
			//fmt.Printf("ABRS %02d: wrn=%q\n", i, tt.want)
			//fmt.Printf("EQUAL = %v %v\n", len(got) == 0, len(tt.want) == 0)
			if len(got) == 0 && len(tt.want) == 0 {
				// assert.Equal fails to notice this is equal.
				// TODO(tlim): Re-check this once everything else works.
			} else {
				if assert.Equal(t, got, tt.want, "Oh no!!!") {
					t.Errorf("analyzeByRecordSet() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func Test_diffTargets(t *testing.T) {
	type args struct {
		existing []targetConfig
		desired  []targetConfig
	}
	tests := []struct {
		name string
		args args
		want ChangeList
	}{

		{
			name: "single",
			args: args{
				existing: []targetConfig{
					{compareable: "1.2.3.4", rec: e0},
				},
				desired: []targetConfig{
					{compareable: "1.2.3.4", rec: e0},
				},
			},
			//want: ,
		},

		{
			name: "add1",
			args: args{
				existing: []targetConfig{
					{compareable: "1.2.3.4", rec: e0},
				},
				desired: []targetConfig{
					{compareable: "1.2.3.4", rec: e0},
					{compareable: "10 laba", rec: e1},
				},
			},
			want: ChangeList{
				Change{Type: CREATE,
					Key:  models.RecordKey{NameFQDN: "laba.f.com", Type: "MX"},
					New:  models.Records{e1},
					Msgs: []string{"CREATE laba.f.com MX 10 laba"},
				},
			},
		},

		{
			name: "del1",
			args: args{
				existing: []targetConfig{
					{compareable: "1.2.3.4", rec: e0},
					{compareable: "10 laba", rec: e1},
				},
				desired: []targetConfig{
					{compareable: "1.2.3.4", rec: e0},
				},
			},
			want: ChangeList{
				Change{Type: DELETE,
					Key:  models.RecordKey{NameFQDN: "laba.f.com", Type: "MX"},
					Old:  models.Records{e1},
					Msgs: []string{"DELETE laba.f.com MX 10 laba"},
				},
			},
		},

		{
			name: "change2nd",
			args: args{
				existing: []targetConfig{
					{compareable: "1.2.3.4", rec: e0},
					{compareable: "10 laba", rec: e1},
				},
				desired: []targetConfig{
					{compareable: "1.2.3.4", rec: e0},
					{compareable: "20 laba", rec: d2},
				},
			},
			want: ChangeList{
				Change{Type: CHANGE,
					Key:  models.RecordKey{NameFQDN: "laba.f.com", Type: "MX"},
					Old:  models.Records{e1},
					New:  models.Records{d2},
					Msgs: []string{"CHANGE laba.f.com MX 20 labb"},
				},
			},
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("DEBUG: Test %02d\n", i)
			if got := diffTargets(tt.args.existing, tt.args.desired); !reflect.DeepEqual(got, tt.want) {
				fmt.Printf("DEBUG: %d %d\n", len(got), len(tt.want))
				t.Errorf("diffTargets()\n GOT=%+v\nWANT=%+v", got, tt.want)
			}
		})
	}
}

func Test_removeCommon(t *testing.T) {
	type args struct {
		existing []targetConfig
		desired  []targetConfig
	}
	tests := []struct {
		name  string
		args  args
		want  []targetConfig
		want1 []targetConfig
	}{
		{
			name: "same",
			args: args{
				existing: []targetConfig{d0tc},
				desired:  []targetConfig{d0tc},
			},
			want:  []targetConfig{},
			want1: []targetConfig{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := removeCommon(tt.args.existing, tt.args.desired)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeCommon() got  = %v, want %v", got, tt.want)
			}
			if (!(got1 == nil && tt.want1 == nil)) && !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("removeCommon() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
