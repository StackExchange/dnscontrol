package natsort

import "testing"

func TestSort(t *testing.T) {

	tests := []struct {
		d1, d2 string
		e1     bool
	}{
		{``, ``, false},
		{`foo3`, `foo10`, true},
		{`foo10`, `foo3`, false},
		{`foo2`, `foo40`, true},
		{`foo40`, `foo2`, false},
		{`ny-dc01.ds`, `ny-dc-vpn`, true},
		{`20161108174726pm._domainkey`, `*`, true},
		{`co-dc01.ds.stackexchange.com`, `co-dc-vpn.stackexchange.com`, true},
	}
	for i, test := range tests {
		r := Less(test.d1, test.d2)
		if r != test.e1 {
			t.Errorf("%v: expected (%v) got (%v): (%v) < (%v)", i, test.e1, r, test.d1, test.d2)
		}
	}
}
