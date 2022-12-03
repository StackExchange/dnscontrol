package diff2

import (
	"reflect"
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
)

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

func Test_analyzeByRecordSet(t *testing.T) {

	origin := "f.com"

	e0 := makeRec(origin, "laba", "A", "1.2.3.4")       //      [0]
	e1 := makeRec(origin, "laba", "MX", "10 laba")      //     [1]
	e2 := makeRec(origin, "labc", "CNAME", "laba")      //     [2]
	e3 := makeRec(origin, "labe", "A", "10.10.10.15")   //  [3]
	e4 := makeRec(origin, "labe", "A", "10.10.10.16")   //  [4]
	e5 := makeRec(origin, "labe", "A", "10.10.10.17")   //  [5]
	e6 := makeRec(origin, "labe", "A", "10.10.10.18")   //  [6]
	e7 := makeRec(origin, "labg", "NS", "10.10.10.15")  // [7]
	e8 := makeRec(origin, "labg", "NS", "10.10.10.16")  // [8]
	e9 := makeRec(origin, "labg", "NS", "10.10.10.17")  // [9]
	e10 := makeRec(origin, "labg", "NS", "10.10.10.18") // [10]
	e11 := makeRec(origin, "labh", "CNAME", "labd")     //     [11]

	existing := models.Records{e0, e1, e2, e3, e4, e5, e6, e7, e8, e9, e9, e10, e11}

	d0 := makeRec(origin, "laba", "A", "1.2.3.4")       //      [0']
	d1 := makeRec(origin, "laba", "A", "1.2.3.5")       //      [1']
	d2 := makeRec(origin, "laba", "MX", "20 labb")      //     [2']
	d3 := makeRec(origin, "labe", "A", "10.10.10.95")   //  [3']
	d4 := makeRec(origin, "labe", "A", "10.10.10.96")   //  [4']
	d5 := makeRec(origin, "labe", "A", "10.10.10.97")   //  [5']
	d6 := makeRec(origin, "labe", "A", "10.10.10.98")   //  [6']
	d7 := makeRec(origin, "labf", "TXT", "foo")         //      [7']
	d8 := makeRec(origin, "labg", "NS", "10.10.10.10")  // [8']
	d9 := makeRec(origin, "labg", "NS", "10.10.10.15")  // [9']
	d10 := makeRec(origin, "labg", "NS", "10.10.10.16") // [10']
	d11 := makeRec(origin, "labg", "NS", "10.10.10.97") // [11']
	d12 := makeRec(origin, "labh", "A", "1.2.3.4")      //      [12']
	desired := models.Records{d0, d1, d2, d3, d4, d5, d6, d7, d8, d9, d10, d11, d12}

	cc := NewCompareConfig(origin, existing, desired, nil)

	want := makeChange(CHANGE, "laba.f.com", "A", nil, models.Records{e0, d1}, []string{"CHANGE"})

	t.Run("01", func(t *testing.T) {
		if got := analyzeByRecordSet(cc); !reflect.DeepEqual(got, want) {
			t.Errorf("analyzeByRecordSet() = %v, want %v", got, want)
		}
	})
}
