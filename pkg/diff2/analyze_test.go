package diff2

import (
	"reflect"
	"strings"
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/kylelemons/godebug/diff"
)

var testDataAA1234 = makeRec("laba", "A", "1.2.3.4")   //      [0]
var testDataAA5678 = makeRec("laba", "A", "5.6.7.8")   //      [0]
var testDataAMX10a = makeRec("laba", "MX", "10 laba")  //     [1]
var testDataCCa = makeRec("labc", "CNAME", "laba")     //     [2]
var testDataEA15 = makeRec("labe", "A", "10.10.10.15") //  [3]
var e4 = makeRec("labe", "A", "10.10.10.16")           //  [4]
var e5 = makeRec("labe", "A", "10.10.10.17")           //  [5]
var e6 = makeRec("labe", "A", "10.10.10.18")           //  [6]
var e7 = makeRec("labg", "NS", "10.10.10.15")          // [7]
var e8 = makeRec("labg", "NS", "10.10.10.16")          // [8]
var e9 = makeRec("labg", "NS", "10.10.10.17")          // [9]
var e10 = makeRec("labg", "NS", "10.10.10.18")         // [10]
var e11mx = makeRec("labh", "MX", "22 ttt")            //     [11]
var e11 = makeRec("labh", "CNAME", "labd")             //     [11]
var testDataApexMX1aaa = makeRec("", "MX", "1 aaa")

var testDataAA1234clone = makeRec("laba", "A", "1.2.3.4") //      [0']
var testDataAA12345 = makeRec("laba", "A", "1.2.3.5")     //      [1']
var testDataAMX20b = makeRec("laba", "MX", "20 labb")     //     [2']
var d3 = makeRec("labe", "A", "10.10.10.95")              //  [3']
var d4 = makeRec("labe", "A", "10.10.10.96")              //  [4']
var d5 = makeRec("labe", "A", "10.10.10.97")              //  [5']
var d6 = makeRec("labe", "A", "10.10.10.98")              //  [6']
var d7 = makeRec("labf", "TXT", "foo")                    //      [7']
var d8 = makeRec("labg", "NS", "10.10.10.10")             // [8']
var d9 = makeRec("labg", "NS", "10.10.10.15")             // [9']
var d10 = makeRec("labg", "NS", "10.10.10.16")            // [10']
var d11 = makeRec("labg", "NS", "10.10.10.97")            // [11']
var d12 = makeRec("labh", "A", "1.2.3.4")                 //      [12']
var testDataApexMX22bbb = makeRec("", "MX", "22 bbb")

var d0tc = targetConfig{compareable: "1.2.3.4 ttl=0", rec: testDataAA1234clone}

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

func compareMsgs(t *testing.T, fnname, testname, testpart string, gotcc ChangeList, wantstring string) {
	t.Helper()
	gs := strings.TrimSpace(justMsgString(gotcc))
	ws := strings.TrimSpace(wantstring)
	d := diff.Diff(gs, ws)
	if d != "" {
		t.Errorf("%s()/%s (wantMsgs:%s):\n===got===\n%s\n===want===\n%s\n===diff===\n%s\n===", fnname, testname, testpart, gs, ws, d)
	}
}

func compareCL(t *testing.T, fnname, testname, testpart string, gotcl ChangeList, wantstring string) {
	t.Helper()
	gs := strings.TrimSpace(gotcl.String())
	ws := strings.TrimSpace(wantstring)
	d := diff.Diff(gs, ws)
	if d != "" {
		t.Errorf("%s()/%s (wantChange%s):\n===got===\n%s\n===want===\n%s\n===diff===\n%s\n===", fnname, testname, testpart, gs, ws, d)
	}

}

func Test_analyzeByRecordSet(t *testing.T) {
	type args struct {
		origin            string
		existing, desired models.Records
		compFn            ComparableFunc
	}

	origin := "f.com"
	existing := models.Records{testDataAA1234, testDataAMX10a, testDataCCa, testDataEA15, e4, e5, e6, e7, e8, e9, e10, e11}
	desired := models.Records{testDataAA1234clone, testDataAA12345, testDataAMX20b, d3, d4, d5, d6, d7, d8, d9, d10, d11, d12}

	tests := []struct {
		name            string
		args            args
		wantMsgs        string
		wantChangeRSet  string
		wantChangeLabel string
		wantChangeRec   string
		wantChangeZone  string
	}{

		{
			name: "oneequal",
			args: args{
				origin:   origin,
				existing: models.Records{testDataAA1234},
				desired:  models.Records{testDataAA1234clone},
			},
			wantMsgs:        "", // Empty
			wantChangeRSet:  "ChangeList: len=0",
			wantChangeLabel: "ChangeList: len=0",
			wantChangeRec:   "ChangeList: len=0",
		},

		{
			name: "onediff",
			args: args{
				origin:   origin,
				existing: models.Records{testDataAA1234, testDataAMX10a},
				desired:  models.Records{testDataAA1234clone, testDataAMX20b},
			},
			wantMsgs: "CHANGE laba.f.com MX (10 laba) -> (20 labb)",
			wantChangeRSet: `
ChangeList: len=1
00: Change: verb=CHANGE
    key={laba.f.com MX}
    old=[10 laba]
    new=[20 labb]
    msg=["CHANGE laba.f.com MX (10 laba) -> (20 labb)"]
`,
			wantChangeLabel: `
ChangeList: len=1
00: Change: verb=CHANGE
    key={laba.f.com }
    old=[1.2.3.4 10 laba]
    new=[1.2.3.4 20 labb]
    msg=["CHANGE laba.f.com MX (10 laba) -> (20 labb)"]
`,
			wantChangeRec: `
ChangeList: len=1
00: Change: verb=CHANGE
    key={laba.f.com MX}
    old=[10 laba]
    new=[20 labb]
    msg=["CHANGE laba.f.com MX (10 laba) -> (20 labb)"]
`,
		},

		{
			name: "apex",
			args: args{
				origin:   origin,
				existing: models.Records{testDataAA1234, testDataApexMX1aaa},
				desired:  models.Records{testDataAA1234clone, testDataApexMX22bbb},
			},
			wantMsgs: "CHANGE f.com MX (1 aaa) -> (22 bbb)",
			wantChangeRSet: `
ChangeList: len=1
00: Change: verb=CHANGE
    key={f.com MX}
    old=[1 aaa]
    new=[22 bbb]
    msg=["CHANGE f.com MX (1 aaa) -> (22 bbb)"]
`,
			wantChangeLabel: `
ChangeList: len=1
00: Change: verb=CHANGE
    key={f.com }
    old=[1 aaa]
    new=[22 bbb]
    msg=["CHANGE f.com MX (1 aaa) -> (22 bbb)"]
`,
			wantChangeRec: `
ChangeList: len=1
00: Change: verb=CHANGE
    key={f.com MX}
    old=[1 aaa]
    new=[22 bbb]
    msg=["CHANGE f.com MX (1 aaa) -> (22 bbb)"]
`,
		},

		{
			name: "firsta",
			args: args{
				origin:   origin,
				existing: models.Records{testDataAA1234, testDataAMX10a},
				desired:  models.Records{testDataAA1234clone, testDataAA12345, testDataAMX20b},
			},
			wantMsgs: `
CREATE laba.f.com A 1.2.3.5
CHANGE laba.f.com MX (10 laba) -> (20 labb)
		`,
			wantChangeRSet: `
ChangeList: len=2
00: Change: verb=CHANGE
    key={laba.f.com A}
    old=[1.2.3.4]
    new=[1.2.3.4 1.2.3.5]
    msg=["CREATE laba.f.com A 1.2.3.5"]
01: Change: verb=CHANGE
    key={laba.f.com MX}
    old=[10 laba]
    new=[20 labb]
    msg=["CHANGE laba.f.com MX (10 laba) -> (20 labb)"]
		`,
			wantChangeLabel: `
ChangeList: len=1
00: Change: verb=CHANGE
    key={laba.f.com }
    old=[1.2.3.4 10 laba]
    new=[1.2.3.4 1.2.3.5 20 labb]
    msg=["CREATE laba.f.com A 1.2.3.5" "CHANGE laba.f.com MX (10 laba) -> (20 labb)"]
		`,
			wantChangeRec: `
ChangeList: len=2
00: Change: verb=CREATE
    key={laba.f.com A}
    new=[1.2.3.5]
    msg=["CREATE laba.f.com A 1.2.3.5"]
01: Change: verb=CHANGE
    key={laba.f.com MX}
    old=[10 laba]
    new=[20 labb]
    msg=["CHANGE laba.f.com MX (10 laba) -> (20 labb)"]
`,
		},

		{
			name: "big",
			args: args{
				origin:   origin,
				existing: existing,
				desired:  desired,
			},
			wantMsgs: `
CREATE laba.f.com A 1.2.3.5
CHANGE laba.f.com MX (10 laba) -> (20 labb)
DELETE labc.f.com CNAME laba
CHANGE labe.f.com A (10.10.10.15) -> (10.10.10.95)
CHANGE labe.f.com A (10.10.10.16) -> (10.10.10.96)
CHANGE labe.f.com A (10.10.10.17) -> (10.10.10.97)
CHANGE labe.f.com A (10.10.10.18) -> (10.10.10.98)
CREATE labf.f.com TXT "foo"
CHANGE labg.f.com NS (10.10.10.17) -> (10.10.10.10)
CHANGE labg.f.com NS (10.10.10.18) -> (10.10.10.97)
DELETE labh.f.com CNAME labd
CREATE labh.f.com A 1.2.3.4
		`,
			wantChangeRSet: `
ChangeList: len=8
00: Change: verb=CHANGE
    key={laba.f.com A}
    old=[1.2.3.4]
    new=[1.2.3.4 1.2.3.5]
    msg=["CREATE laba.f.com A 1.2.3.5"]
01: Change: verb=CHANGE
    key={laba.f.com MX}
    old=[10 laba]
    new=[20 labb]
    msg=["CHANGE laba.f.com MX (10 laba) -> (20 labb)"]
02: Change: verb=DELETE
    key={labc.f.com CNAME}
    msg=["DELETE labc.f.com CNAME laba"]
03: Change: verb=CHANGE
    key={labe.f.com A}
    old=[10.10.10.15 10.10.10.16 10.10.10.17 10.10.10.18]
    new=[10.10.10.95 10.10.10.96 10.10.10.97 10.10.10.98]
    msg=["CHANGE labe.f.com A (10.10.10.15) -> (10.10.10.95)" "CHANGE labe.f.com A (10.10.10.16) -> (10.10.10.96)" "CHANGE labe.f.com A (10.10.10.17) -> (10.10.10.97)" "CHANGE labe.f.com A (10.10.10.18) -> (10.10.10.98)"]
04: Change: verb=CREATE
    key={labf.f.com TXT}
    new=["foo"]
    msg=["CREATE labf.f.com TXT \"foo\""]
05: Change: verb=CHANGE
    key={labg.f.com NS}
    old=[10.10.10.15 10.10.10.16 10.10.10.17 10.10.10.18]
    new=[10.10.10.10 10.10.10.15 10.10.10.16 10.10.10.97]
    msg=["CHANGE labg.f.com NS (10.10.10.17) -> (10.10.10.10)" "CHANGE labg.f.com NS (10.10.10.18) -> (10.10.10.97)"]
06: Change: verb=DELETE
    key={labh.f.com CNAME}
    msg=["DELETE labh.f.com CNAME labd"]
07: Change: verb=CREATE
    key={labh.f.com A}
    new=[1.2.3.4]
    msg=["CREATE labh.f.com A 1.2.3.4"]
		`,
			wantChangeLabel: `
ChangeList: len=6
00: Change: verb=CHANGE
    key={laba.f.com }
    old=[1.2.3.4 10 laba]
    new=[1.2.3.4 1.2.3.5 20 labb]
    msg=["CREATE laba.f.com A 1.2.3.5" "CHANGE laba.f.com MX (10 laba) -> (20 labb)"]
01: Change: verb=DELETE
    key={labc.f.com }
    msg=["DELETE labc.f.com CNAME laba"]
02: Change: verb=CHANGE
    key={labe.f.com }
    old=[10.10.10.15 10.10.10.16 10.10.10.17 10.10.10.18]
    new=[10.10.10.95 10.10.10.96 10.10.10.97 10.10.10.98]
    msg=["CHANGE labe.f.com A (10.10.10.15) -> (10.10.10.95)" "CHANGE labe.f.com A (10.10.10.16) -> (10.10.10.96)" "CHANGE labe.f.com A (10.10.10.17) -> (10.10.10.97)" "CHANGE labe.f.com A (10.10.10.18) -> (10.10.10.98)"]
03: Change: verb=CREATE
    key={labf.f.com }
    new=["foo"]
    msg=["CREATE labf.f.com TXT \"foo\""]
04: Change: verb=CHANGE
    key={labg.f.com }
    old=[10.10.10.15 10.10.10.16 10.10.10.17 10.10.10.18]
    new=[10.10.10.10 10.10.10.15 10.10.10.16 10.10.10.97]
    msg=["CHANGE labg.f.com NS (10.10.10.17) -> (10.10.10.10)" "CHANGE labg.f.com NS (10.10.10.18) -> (10.10.10.97)"]
05: Change: verb=CHANGE
    key={labh.f.com }
    old=[labd]
    new=[1.2.3.4]
    msg=["DELETE labh.f.com CNAME labd" "CREATE labh.f.com A 1.2.3.4"]
		`,
			wantChangeRec: `
ChangeList: len=12
00: Change: verb=CREATE
    key={laba.f.com A}
    new=[1.2.3.5]
    msg=["CREATE laba.f.com A 1.2.3.5"]
01: Change: verb=CHANGE
    key={laba.f.com MX}
    old=[10 laba]
    new=[20 labb]
    msg=["CHANGE laba.f.com MX (10 laba) -> (20 labb)"]
02: Change: verb=DELETE
    key={labc.f.com CNAME}
    old=[laba]
    msg=["DELETE labc.f.com CNAME laba"]
03: Change: verb=CHANGE
    key={labe.f.com A}
    old=[10.10.10.15]
    new=[10.10.10.95]
    msg=["CHANGE labe.f.com A (10.10.10.15) -> (10.10.10.95)"]
04: Change: verb=CHANGE
    key={labe.f.com A}
    old=[10.10.10.16]
    new=[10.10.10.96]
    msg=["CHANGE labe.f.com A (10.10.10.16) -> (10.10.10.96)"]
05: Change: verb=CHANGE
    key={labe.f.com A}
    old=[10.10.10.17]
    new=[10.10.10.97]
    msg=["CHANGE labe.f.com A (10.10.10.17) -> (10.10.10.97)"]
06: Change: verb=CHANGE
    key={labe.f.com A}
    old=[10.10.10.18]
    new=[10.10.10.98]
    msg=["CHANGE labe.f.com A (10.10.10.18) -> (10.10.10.98)"]
07: Change: verb=CREATE
    key={labf.f.com TXT}
    new=["foo"]
    msg=["CREATE labf.f.com TXT \"foo\""]
08: Change: verb=CHANGE
    key={labg.f.com NS}
    old=[10.10.10.17]
    new=[10.10.10.10]
    msg=["CHANGE labg.f.com NS (10.10.10.17) -> (10.10.10.10)"]
09: Change: verb=CHANGE
    key={labg.f.com NS}
    old=[10.10.10.18]
    new=[10.10.10.97]
    msg=["CHANGE labg.f.com NS (10.10.10.18) -> (10.10.10.97)"]
10: Change: verb=DELETE
    key={labh.f.com CNAME}
    old=[labd]
    msg=["DELETE labh.f.com CNAME labd"]
11: Change: verb=CREATE
    key={labh.f.com A}
    new=[1.2.3.4]
    msg=["CREATE labh.f.com A 1.2.3.4"]
`,
		},
	}
	for _, tt := range tests {

		// Each "analyze*()" should return the same msgs, but a different ChangeList.
		// Sadly the analyze*() functions are destructive to the CompareConfig struct.
		// Therefore we have to run NewCompareConfig() each time.

		t.Run(tt.name, func(t *testing.T) {
			cl := analyzeByRecordSet(NewCompareConfig(tt.args.origin, tt.args.existing, tt.args.desired, tt.args.compFn))
			compareMsgs(t, "analyzeByRecordSet", tt.name, "RSet", cl, tt.wantMsgs)
			compareCL(t, "analyzeByRecordSet", tt.name, "RSet", cl, tt.wantChangeRSet)
		})

		t.Run(tt.name, func(t *testing.T) {
			cl := analyzeByLabel(NewCompareConfig(tt.args.origin, tt.args.existing, tt.args.desired, tt.args.compFn))
			compareMsgs(t, "analyzeByLabel", tt.name, "Label", cl, tt.wantMsgs)
			compareCL(t, "analyzeByLabel", tt.name, "Label", cl, tt.wantChangeLabel)
		})

		t.Run(tt.name, func(t *testing.T) {
			cl := analyzeByRecord(NewCompareConfig(tt.args.origin, tt.args.existing, tt.args.desired, tt.args.compFn))
			compareMsgs(t, "analyzeByRecord", tt.name, "Rec", cl, tt.wantMsgs)
			compareCL(t, "analyzeByRecord", tt.name, "Rec", cl, tt.wantChangeRec)
		})

		// NB(tlim): There is no analyzeByZone().  ByZone() uses analyzeByRecord().

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
					{compareable: "1.2.3.4", rec: testDataAA1234},
				},
				desired: []targetConfig{
					{compareable: "1.2.3.4", rec: testDataAA1234},
				},
			},
			//want: ,
		},

		{
			name: "add1",
			args: args{
				existing: []targetConfig{
					{compareable: "1.2.3.4", rec: testDataAA1234},
				},
				desired: []targetConfig{
					{compareable: "1.2.3.4", rec: testDataAA1234},
					{compareable: "10 laba", rec: testDataAMX10a},
				},
			},
			want: ChangeList{
				Change{Type: CREATE,
					Key:  models.RecordKey{NameFQDN: "laba.f.com", Type: "MX"},
					New:  models.Records{testDataAMX10a},
					Msgs: []string{"CREATE laba.f.com MX 10 laba"},
				},
			},
		},

		{
			name: "del1",
			args: args{
				existing: []targetConfig{
					{compareable: "1.2.3.4", rec: testDataAA1234},
					{compareable: "10 laba", rec: testDataAMX10a},
				},
				desired: []targetConfig{
					{compareable: "1.2.3.4", rec: testDataAA1234},
				},
			},
			want: ChangeList{
				Change{Type: DELETE,
					Key:  models.RecordKey{NameFQDN: "laba.f.com", Type: "MX"},
					Old:  models.Records{testDataAMX10a},
					Msgs: []string{"DELETE laba.f.com MX 10 laba"},
				},
			},
		},

		{
			name: "change2nd",
			args: args{
				existing: []targetConfig{
					{compareable: "1.2.3.4", rec: testDataAA1234},
					{compareable: "10 laba", rec: testDataAMX10a},
				},
				desired: []targetConfig{
					{compareable: "1.2.3.4", rec: testDataAA1234},
					{compareable: "20 laba", rec: testDataAMX20b},
				},
			},
			want: ChangeList{
				Change{Type: CHANGE,
					Key:  models.RecordKey{NameFQDN: "laba.f.com", Type: "MX"},
					Old:  models.Records{testDataAMX10a},
					New:  models.Records{testDataAMX20b},
					Msgs: []string{"CHANGE laba.f.com MX (10 laba) -> (20 labb)"},
				},
			},
		},

		{
			name: "del2nd",
			args: args{
				existing: []targetConfig{
					{compareable: "1.2.3.4", rec: testDataAA1234},
					{compareable: "5.6.7.8", rec: testDataAA5678},
				},
				desired: []targetConfig{
					{compareable: "1.2.3.4", rec: testDataAA1234},
				},
			},
			want: ChangeList{
				Change{Type: CHANGE,
					Key:  models.RecordKey{NameFQDN: "laba.f.com", Type: "A"},
					Old:  models.Records{testDataAA1234, testDataAA5678},
					New:  models.Records{testDataAA1234},
					Msgs: []string{"DELETE laba.f.com A 5.6.7.8"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//fmt.Printf("DEBUG: Test %02d\n", i)
			got := diffTargets(tt.args.existing, tt.args.desired)
			d := diff.Diff(strings.TrimSpace(justMsgString(got)), strings.TrimSpace(justMsgString(tt.want)))
			if d != "" {
				//fmt.Printf("DEBUG: %d %d\n", len(got), len(tt.want))
				t.Errorf("diffTargets()\n diff=%s", d)
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

func comparables(s []targetConfig) []string {
	var r []string
	for _, j := range s {
		r = append(r, j.compareable)
	}
	return r
}

func Test_filterBy(t *testing.T) {
	type args struct {
		s []targetConfig
		m map[string]*targetConfig
	}
	tests := []struct {
		name string
		args args
		want []targetConfig
	}{

		{
			name: "removeall",
			args: args{
				s: []targetConfig{
					{compareable: "1.2.3.4", rec: testDataAA1234},
					{compareable: "10 laba", rec: testDataAMX10a},
				},
				m: map[string]*targetConfig{
					"1.2.3.4": {compareable: "1.2.3.4", rec: testDataAA1234},
					"10 laba": {compareable: "10 laba", rec: testDataAMX10a},
				},
			},
			want: []targetConfig{},
		},

		{
			name: "keepall",
			args: args{
				s: []targetConfig{
					{compareable: "1.2.3.4", rec: testDataAA1234},
					{compareable: "10 laba", rec: testDataAMX10a},
				},
				m: map[string]*targetConfig{
					"nothing": {compareable: "1.2.3.4", rec: testDataAA1234},
					"matches": {compareable: "10 laba", rec: testDataAMX10a},
				},
			},
			want: []targetConfig{
				{compareable: "1.2.3.4", rec: testDataAA1234},
				{compareable: "10 laba", rec: testDataAMX10a},
			},
		},

		{
			name: "keepsome",
			args: args{
				s: []targetConfig{
					{compareable: "1.2.3.4", rec: testDataAA1234},
					{compareable: "10 laba", rec: testDataAMX10a},
				},
				m: map[string]*targetConfig{
					"nothing": {compareable: "1.2.3.4", rec: testDataAA1234},
					"10 laba": {compareable: "10 laba", rec: testDataAMX10a},
				},
			},
			want: []targetConfig{
				{compareable: "1.2.3.4", rec: testDataAA1234},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterBy(tt.args.s, tt.args.m); !reflect.DeepEqual(comparables(got), comparables(tt.want)) {
				t.Errorf("filterBy() = %v, want %v", got, tt.want)
			}
		})
	}
}
