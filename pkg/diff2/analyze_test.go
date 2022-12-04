package diff2

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
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

// func Test_analyzeByRecordSet(t *testing.T) {

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
					targetConfig{compareable: "1.2.3.4", rec: e0},
				},
				desired: []targetConfig{
					targetConfig{compareable: "1.2.3.4", rec: e0},
				},
			},
			//want: ,
		},

		{
			name: "add1",
			args: args{
				existing: []targetConfig{
					targetConfig{compareable: "1.2.3.4", rec: e0},
				},
				desired: []targetConfig{
					targetConfig{compareable: "1.2.3.4", rec: e0},
					targetConfig{compareable: "10 laba", rec: e1},
				},
			},
			want: ChangeList{
				Change{Type: CREATE,
					Key:  models.RecordKey{"laba.f.com", "MX"},
					New:  models.Records{e1},
					Msgs: []string{"CREATE laba.f.com MX 10 laba"},
				},
			},
		},

		{
			name: "del1",
			args: args{
				existing: []targetConfig{
					targetConfig{compareable: "1.2.3.4", rec: e0},
					targetConfig{compareable: "10 laba", rec: e1},
				},
				desired: []targetConfig{
					targetConfig{compareable: "1.2.3.4", rec: e0},
				},
			},
			want: ChangeList{
				Change{Type: DELETE,
					Key:  models.RecordKey{"laba.f.com", "MX"},
					Old:  models.Records{e1},
					Msgs: []string{"DELETE laba.f.com MX 10 laba"},
				},
			},
		},

		{
			name: "change1",
			args: args{
				existing: []targetConfig{
					targetConfig{compareable: "1.2.3.4", rec: e0},
					targetConfig{compareable: "10 laba", rec: e1},
				},
				desired: []targetConfig{
					targetConfig{compareable: "1.2.3.4", rec: e0},
					targetConfig{compareable: "20 laba", rec: d2},
				},
			},
			want: ChangeList{
				Change{Type: CHANGE,
					Key:  models.RecordKey{"laba.f.com", "MX"},
					Old:  models.Records{e0, e1},
					New:  models.Records{e0, d2},
					Msgs: []string{"CCC laba.f.com MX 10 laba"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := diffTargets(tt.args.existing, tt.args.desired); !reflect.DeepEqual(got, tt.want) {
				fmt.Printf("DEBUG: %d %d\n", len(got), len(tt.want))
				t.Errorf("diffTargets()\n GOT=%+v\nWANT=%+v", got, tt.want)
			}
		})
	}
}
