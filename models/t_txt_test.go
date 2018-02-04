package models

// import "testing"

// func TestSetTxtParse(t *testing.T) {
// 	tests := []struct {
// 		d1 string
// 		e1 string
// 		e2 []string
// 	}{
// 		{``, ``, []string{``}},
// 		{`foo`, `foo`, []string{`foo`}},
// 		{`"foo"`, `foo`, []string{`foo`}},
// 		{`"foo bar"`, `foo bar`, []string{`foo bar`}},
// 		{`foo bar`, `foo bar`, []string{`foo bar`}},
// 		{`"aaa" "bbb"`, `aaa`, []string{`aaa`, `bbb`}},
// 	}
// 	for i, test := range tests {
// 		x := &RecordConfig{Type: "TXT"}
// 		x.SetTargetTxtParse(test.d1)
// 		if x.Target != test.e1 {
// 			t.Errorf("%v: expected Target=(%v) got (%v)", i, test.e1, x.Target)
// 		}
// 		if len(x.TxtStrings) != len(test.e2) {
// 			t.Errorf("%v: expected TxtStrings=(%v) got (%v)", i, test.e2, x.TxtStrings)
// 		}
// 		for i := range x.TxtStrings {
// 			if len(x.TxtStrings[i]) != len(test.e2[i]) {
// 				t.Errorf("%v: expected TxtStrings=(%v) got (%v)", i, test.e2, x.TxtStrings)
// 			}
// 		}
// 	}
// }
