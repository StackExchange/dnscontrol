package namedotcom

import (
	"strings"
	"testing"
)

var txtData = []struct {
	decoded []string
	encoded string
}{
	{[]string{`simple`}, `simple`},
	{[]string{`changed`}, `changed`},
	{[]string{`with spaces`}, `with spaces`},
	{[]string{`with whitespace`}, `with whitespace`},
	{[]string{"one", "two"}, `"one""two"`},
	{[]string{"eh", "bee", "cee"}, `"eh""bee""cee"`},
	{[]string{"o\"ne", "tw\"o"}, `"o\"ne""tw\"o"`},
	{[]string{"dimple"}, `dimple`},
	//{[]string{`back\slash`}, `"back\\slash"`}, // Not yet supported
	//{[]string{`back\\slash`}, `"back\\\\slash"`}, // Not yet supported
	//{[]string{`back\\\slash`}, `"back\\\\\\slash"`}, // Not yet supported
	{[]string{"fun", "two"}, `"fun""two"`},
	{[]string{"eh", "bzz", "cee"}, `"eh""bzz""cee"`},
}

func TestDecodeTxt(t *testing.T) {
	// Test decoded a string into the list of strings:
	for i, test := range txtData {
		data := test.encoded
		got := decodeTxt(data)
		wanted := test.decoded
		if len(got) != len(wanted) {
			t.Errorf("%v: txt\n  decode: %v\nexpected: `%v`\n     got: `%v`\n", i, data, strings.Join(wanted, "`, `"), strings.Join(got, "`, `"))
		} else {
			for j := range got {
				if got[j] != wanted[j] {
					t.Errorf("%v: txt\n  decode: %v\nexpected: `%v`\n     got: `%v`\n", i, data, strings.Join(wanted, "`, `"), strings.Join(got, "`, `"))
				}
			}
		}
	}
}
