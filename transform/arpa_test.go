package transform

import "testing"
import "fmt"

func TestReverse(t *testing.T) {
	var tests = []struct {
		in      string
		isError bool
		out     string
	}{
		{"174.136.107.0/24", false, "107.136.174.in-addr.arpa"},

		{"174.136.0.0/16", false, "136.174.in-addr.arpa"},
		{"174.136.43.0/16", false, "136.174.in-addr.arpa"}, //do bits set inside the masked range matter? Should this be invalid? Is there a shorter way to specify this?

		{"174.0.0.0/8", false, "174.in-addr.arpa"},
		{"174.136.43.0/8", false, "174.in-addr.arpa"},
		{"174.136.43.0/8", false, "174.in-addr.arpa"},

		{"2001::/16", false, "1.0.0.2.ip6.arpa"},
		{"2001:0db8:0123:4567:89ab:cdef:1234:5670/124", false, "7.6.5.4.3.2.1.f.e.d.c.b.a.9.8.7.6.5.4.3.2.1.0.8.b.d.0.1.0.0.2.ip6.arpa"},

		{"174.136.107.14/32", false, "14.107.136.174.in-addr.arpa"},
		{"2001:0db8:0123:4567:89ab:cdef:1234:5678/128", false, "8.7.6.5.4.3.2.1.f.e.d.c.b.a.9.8.7.6.5.4.3.2.1.0.8.b.d.0.1.0.0.2.ip6.arpa"},

		//Errror Cases:
		{"0.0.0.0/0", true, ""},
		{"2001::/0", true, ""},
		{"4.5/16", true, ""},
		{"foo.com", true, ""},
	}
	for i, tst := range tests {
		t.Run(fmt.Sprintf("%d--%s", i, tst.in), func(t *testing.T) {
			d, err := ReverseDomainName(tst.in)
			if err != nil && !tst.isError {
				t.Error("Should not have errored ", err)
			} else if tst.isError && err == nil {
				t.Errorf("Should have errored, but didn't. Got %s", d)
			} else if d != tst.out {
				t.Errorf("Expected '%s' but got '%s'", tst.out, d)
			}
		})
	}
}
