package natsort

import "testing"

func TestLess(t *testing.T) {

	tests := []struct {
		d1, d2 string
	}{

		{``, `x`},
		{`foo3`, `foo10`},
		{`foo2`, `foo40`},
		{`ny-dc01.ds`, `ny-dc-vpn`},
		{`20161108174726pm._domainkey`, `*`},
		{`co-dc01.ds.stackexchange.com`, `co-dc-vpn.stackexchange.com`},

		// Bug#4: [a1 a#1 a_1 aa]
		//{`a1`, `a#1`},
		{`a#1`, `a_1`},
		{`a_1`, `aa`},

		// Bug#4: [1 #1 _1 a]
		//{`1`, `#1`},  // Not implemented
		{`#1`, `_1`},
		{`_1`, `a`},

		// Bug#4: [1 01 2 02 3]
		{`1`, `01`},
		{`2`, `02`},
		//{`02`, `3`}, // Not implemented

		// Bug#4: [111111111111111111112 111111111111111111113 1111111111111111111120]
		{`111111111111111111112`, `111111111111111111113`},
		{`111111111111111111113`, `1111111111111111111120`},
	}
	for i, test := range tests {
		r1 := Less(test.d1, test.d2)
		if r1 != true {
			t.Errorf("%va: expected (%v) got (%v): (%v) < (%v)", i, true, r1, test.d1, test.d2)
		} else {
			r2 := Less(test.d2, test.d1)
			if r2 == true {
				t.Errorf("%vb: expected (%v) got (%v): (%v) < (%v)", i, false, r2, test.d1, test.d2)
			}
		}
	}
}

func TestEqual(t *testing.T) {

	tests := []struct {
		d1, d2 string
	}{
		{``, ``},
	}
	for i, test := range tests {
		r1 := Less(test.d1, test.d2)
		if r1 != false {
			t.Errorf("%va: expected (%v) got (%v): (%v) < (%v)", i, false, r1, test.d1, test.d2)
		} else {
			r2 := Less(test.d2, test.d1)
			if r2 != false {
				t.Errorf("%vb: expected (%v) got (%v): (%v) < (%v)", i, false, r2, test.d1, test.d2)
			}
		}
	}
}
