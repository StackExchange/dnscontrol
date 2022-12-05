package diff2

import (
	"encoding/json"
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/kylelemons/godebug/diff"
)

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func TestNewCompareConfig(t *testing.T) {
	type args struct {
		origin   string
		existing models.Records
		desired  models.Records
		compFn   ComparableFunc
	}
	tests := []struct {
		name string
		args args
		want string
	}{

		{
			name: "01",
			args: args{
				origin:   "f.com",
				existing: models.Records{e0},
				desired:  models.Records{d0},
				compFn:   nil,
			},
			want: `existing: [1.2.3.4]
desired: [1.2.3.4]
ldata:
  ldata[00]: laba.f.com
    tdata[0]: "A" (1, 1, 1, 1)
origin: f.com
compFn: <nil>
labelMap: map[laba.f.com:true]
keyMap:   map[{laba.f.com A}:true]`,
			// want: &CompareConfig{
			// 	existing: models.Records{e0},
			// 	desired:  models.Records{d0},
			// 	ldata: []*labelConfig{
			// 		{
			// 			label: "laba.f.com",
			// 			tdata: []*rTypeConfig{
			// 				{
			// 					rType: "A",
			// 					existingTargets: []targetConfig{
			// 						{compareable: "1.2.3.4 ttl=0", rec: e0},
			// 					},
			// 					desiredTargets: []targetConfig{
			// 						{compareable: "1.2.3.4 ttl=0", rec: d0},
			// 					},
			// 					existingRecs: []*models.RecordConfig{e0},
			// 					desiredRecs:  []*models.RecordConfig{d0},
			// 				},
			// 			},
			// 		},
			// 	},
			// 	origin:          "f.com",
			// 	compareableFunc: nil,
			// 	labelMap:        map[string]bool{"laba.f.com": true},
			// 	keyMap:          map[models.RecordKey]bool{e0.Key(): true},
			// },
		},

		{
			name: "02",
			args: args{
				origin:   "f.com",
				existing: models.Records{e0, e2},
				desired:  models.Records{d0},
				compFn:   nil,
			},
			want: "bar",
			// want: &CompareConfig{
			// 	existing: models.Records{e0, e2},
			// 	desired:  models.Records{d0},
			// 	ldata: []*labelConfig{
			// 		// {
			// 		// 	label: "laba.f.com",
			// 		// 	tdata: []*rTypeConfig{
			// 		// 		{
			// 		// 			rType: "A",
			// 		// 			existingTargets: []targetConfig{
			// 		// 				{compareable: "1.2.3.4 ttl=0", rec: e0},
			// 		// 			},
			// 		// 			desiredTargets: []targetConfig{
			// 		// 				{compareable: "1.2.3.4 ttl=0", rec: d0},
			// 		// 			},
			// 		// 			existingRecs: []*models.RecordConfig{e0},
			// 		// 			desiredRecs:  []*models.RecordConfig{d0},
			// 		// 		},
			// 		// 	},
			// 		// },
			// 		// {
			// 		// 	label: "labc.f.com",
			// 		// 	tdata: []*rTypeConfig{
			// 		// 		{
			// 		// 			rType: "A",
			// 		// 			existingTargets: []targetConfig{
			// 		// 				{compareable: "1.2.3.4 ttl=0", rec: e0},
			// 		// 				{compareable: "laba ttl=0", rec: e2},
			// 		// 			},
			// 		// 			desiredTargets: []targetConfig{
			// 		// 				{compareable: "1.2.3.4 ttl=0", rec: d0},
			// 		// 			},
			// 		// 			existingRecs: []*models.RecordConfig{e0},
			// 		// 			desiredRecs:  []*models.RecordConfig{d0},
			// 		// 		},
			// 		// 	},
			// 		// },
			// 		{
			// 			label: "laba.f.com",
			// 			tdata: []*rTypeConfig{
			// 				{
			// 					rType: "A",
			// 					existingTargets: []targetConfig{
			// 						{compareable: "10 laba ttl=0", rec: e1},
			// 					},
			// 					desiredTargets: []targetConfig{
			// 						{compareable: "20 labb ttl=0", rec: d2},
			// 					},
			// 					existingRecs: []*models.RecordConfig{e1},
			// 					desiredRecs:  []*models.RecordConfig{d2},
			// 				},
			// 			},
			// 		},
			// 	},
			// 	origin:          "f.com",
			// 	compareableFunc: nil,
			// 	labelMap:        map[string]bool{"laba.f.com": true, "labc.f.com": true},
			// 	keyMap:          map[models.RecordKey]bool{e0.Key(): true, e2.Key(): true},
			// },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//fmt.Printf("TEST %02d\n", i)
			cc := NewCompareConfig(tt.args.origin, tt.args.existing, tt.args.desired, tt.args.compFn)
			got := cc.String()
			if got != tt.want {
				d := diff.Diff(got, tt.want)
				//t.Errorf("NewCompareConfig() = \n GOT:\n%v\nWANT:\n%v", got, tt.want)
				t.Errorf("NewCompareConfig() = \n%s\n", d)
			}
			// if !reflect.DeepEqual(got.ldata, tt.want.ldata) {
			// 	for i, j := range got.ldata {
			// 		fmt.Printf("CC got=%d:%+v\n", i, j)
			// 	}
			// 	for i, j := range tt.want.ldata {
			// 		fmt.Printf("CC wnt=%d:%+v\n", i, j)
			// 	}
			// 	t.Errorf("NewCompareConfig() = ldata DEEP\n got=%v\nwant=%v", got.ldata, tt.want.ldata)
			// }
			// cc := NewCompareConfig(tt.args.origin, tt.args.existing, tt.args.desired, tt.args.compFn)
			// spew.Dump(cc)
			// if compareCC(t, got, tt.want) != nil {
			// 	fmt.Println("GOT:")
			// 	spew.Dump(got)
			// 	fmt.Println("WANT:")
			// 	spew.Dump(tt.want)
			// 	fmt.Println()
			// 	t.Errorf("NewCompareConfig() = %v, want %v", got, tt.want)
			// }
			// if compareCC(t, got, tt.want) != nil {
			// 	t.Errorf("NewCompareConfig() = %v, want %v", got, tt.want)
			// }
			// if got := NewCompareConfig(tt.args.origin, tt.args.existing, tt.args.desired, tt.args.compFn); !reflect.DeepEqual(got, tt.want) {
			// 	fmt.Println("GOT:")
			// 	spew.Dump(got)
			// 	fmt.Println("WANT:")
			// 	spew.Dump(tt.want)
			// 	fmt.Println()
			// 	t.Errorf("NewCompareConfig() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func compareCC(t *testing.T, got, want *CompareConfig) error {

	// cl, _ := structdiff.Diff(got.existing, want.existing)
	// fmt.Printf("DIFF existing=%s\n", cl)
	// cl, _ = structdiff.Diff(got.desired, want.desired)
	// fmt.Printf("DIFF desired=%s\n", cl)
	// cl, _ = structdiff.Diff(got.ldata, want.ldata)
	// fmt.Printf("DIFF ldata=%s\n", cl)
	// cl, _ = structdiff.Diff(got.origin, want.origin)
	// fmt.Printf("DIFF origin=%s\n", cl)
	// cl, _ = structdiff.Diff(got.compareableFunc, want.compareableFunc)
	// fmt.Printf("DIFF compFn=%s\n", cl)
	// cl, _ = structdiff.Diff(got.labelMap, want.labelMap)
	// fmt.Printf("DIFF labelMap=%s\n", cl)
	// cl, _ = structdiff.Diff(got.keyMap, want.keyMap)
	// fmt.Printf("DIFF keyMap=%s\n", cl)

	if len(got.existing) != len(want.existing) {
		t.Fatalf("existing wrong len got=%d want=%d", len(got.existing), len(want.existing))
	}

	if len(got.desired) != len(want.desired) {
		t.Fatalf("desired wrong len got=%d want=%d", len(got.desired), len(want.desired))
	}

	if len(got.ldata) != len(want.ldata) {
		t.Fatalf("ldata wrong len got=%d want=%d", len(got.ldata), len(want.ldata))
	}

	if len(got.labelMap) != len(want.labelMap) {
		t.Fatalf("labelMap wrong len got=%d want=%d", len(got.labelMap), len(want.labelMap))
	}

	if len(got.keyMap) != len(want.keyMap) {
		t.Fatalf("keyMap wrong len got=%d want=%d", len(got.keyMap), len(want.keyMap))
	}

	return nil
}
