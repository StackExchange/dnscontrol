package rfc4183

import (
	"fmt"
	"testing"
)

func TestReverse(t *testing.T) {
	var tests = []struct {
		in  string
		out string
	}{
		// IPv4 "Classless in-addr.arpa delegation" RFC4183.
		// Examples in the RFC:
		{"10.100.2.0/26", "0-26.2.100.10.in-addr.arpa"},
		{"10.192.0.0/13", "192-13.10.in-addr.arpa"},
		{"10.20.128.0/23", "128-23.20.10.in-addr.arpa"},
		{"10.20.129.0/23", "128-23.20.10.in-addr.arpa"}, // Not in the RFC but should be!

		// IPv6
		{"2001::/16", "1.0.0.2.ip6.arpa"},
		{"2001:0db8:0123:4567:89ab:cdef:1234:5670/64", "7.6.5.4.3.2.1.0.8.b.d.0.1.0.0.2.ip6.arpa"},
		{"2001:0db8:0123:4567:89ab:cdef:1234:5670/68", "8.7.6.5.4.3.2.1.0.8.b.d.0.1.0.0.2.ip6.arpa"},
		{"2001:0db8:0123:4567:89ab:cdef:1234:5670/124", "7.6.5.4.3.2.1.f.e.d.c.b.a.9.8.7.6.5.4.3.2.1.0.8.b.d.0.1.0.0.2.ip6.arpa"},
		{"2001:0db8:0123:4567:89ab:cdef:1234:5678/128", "8.7.6.5.4.3.2.1.f.e.d.c.b.a.9.8.7.6.5.4.3.2.1.0.8.b.d.0.1.0.0.2.ip6.arpa"},

		// 8-bit boundaries
		{"174.0.0.0/8", "174.in-addr.arpa"},
		{"174.136.43.0/8", "174.in-addr.arpa"},
		{"174.136.0.44/8", "174.in-addr.arpa"},
		{"174.136.45.45/8", "174.in-addr.arpa"},
		{"174.136.0.0/16", "136.174.in-addr.arpa"},
		{"174.136.43.0/16", "136.174.in-addr.arpa"},
		{"174.136.44.255/16", "136.174.in-addr.arpa"},
		{"174.136.107.0/24", "107.136.174.in-addr.arpa"},
		{"174.136.107.1/24", "107.136.174.in-addr.arpa"},
		{"174.136.107.14/32", "14.107.136.174.in-addr.arpa"},

		// /25 (all cases)
		{"174.1.0.0/25", "0-25.0.1.174.in-addr.arpa"},
		{"174.1.0.128/25", "128-25.0.1.174.in-addr.arpa"},
		{"174.1.0.129/25", "128-25.0.1.174.in-addr.arpa"}, // host bits
		// /26 (all cases)
		{"174.1.0.0/26", "0-26.0.1.174.in-addr.arpa"},
		{"174.1.0.0/26", "0-26.0.1.174.in-addr.arpa"},
		{"174.1.0.64/26", "64-26.0.1.174.in-addr.arpa"},
		{"174.1.0.128/26", "128-26.0.1.174.in-addr.arpa"},
		{"174.1.0.192/26", "192-26.0.1.174.in-addr.arpa"},
		{"174.1.0.194/26", "192-26.0.1.174.in-addr.arpa"}, // host bits
		// /27 (all cases)
		{"174.1.0.0/27", "0-27.0.1.174.in-addr.arpa"},
		{"174.1.0.32/27", "32-27.0.1.174.in-addr.arpa"},
		{"174.1.0.64/27", "64-27.0.1.174.in-addr.arpa"},
		{"174.1.0.96/27", "96-27.0.1.174.in-addr.arpa"},
		{"174.1.0.128/27", "128-27.0.1.174.in-addr.arpa"},
		{"174.1.0.160/27", "160-27.0.1.174.in-addr.arpa"},
		{"174.1.0.192/27", "192-27.0.1.174.in-addr.arpa"},
		{"174.1.0.224/27", "224-27.0.1.174.in-addr.arpa"},
		{"174.1.0.225/27", "224-27.0.1.174.in-addr.arpa"}, // host bits
		// /28 (first 2, last 2)
		{"174.1.0.0/28", "0-28.0.1.174.in-addr.arpa"},
		{"174.1.0.16/28", "16-28.0.1.174.in-addr.arpa"},
		{"174.1.0.224/28", "224-28.0.1.174.in-addr.arpa"},
		{"174.1.0.240/28", "240-28.0.1.174.in-addr.arpa"},
		{"174.1.0.241/28", "240-28.0.1.174.in-addr.arpa"}, // host bits
		// /29 (first 2 cases)
		{"174.1.0.0/29", "0-29.0.1.174.in-addr.arpa"},
		{"174.1.0.8/29", "8-29.0.1.174.in-addr.arpa"},
		{"174.1.0.9/29", "8-29.0.1.174.in-addr.arpa"}, // host bits
		// /30 (first 2 cases)
		{"174.1.0.0/30", "0-30.0.1.174.in-addr.arpa"},
		{"174.1.0.4/30", "4-30.0.1.174.in-addr.arpa"},
		{"174.1.0.5/30", "4-30.0.1.174.in-addr.arpa"}, // host bits
		// /31 (first 2 cases)
		{"174.1.0.0/31", "0-31.0.1.174.in-addr.arpa"},
		{"174.1.0.2/31", "2-31.0.1.174.in-addr.arpa"},
		{"174.1.0.3/31", "2-31.0.1.174.in-addr.arpa"}, // host bits

		// Other tests:
		{"10.100.2.255/23", "2-23.100.10.in-addr.arpa"},
		{"10.100.2.255/22", "0-22.100.10.in-addr.arpa"},
		{"10.100.2.255/21", "0-21.100.10.in-addr.arpa"},
		{"10.100.2.255/20", "0-20.100.10.in-addr.arpa"},
		{"10.100.2.255/19", "0-19.100.10.in-addr.arpa"},
		{"10.100.2.255/18", "0-18.100.10.in-addr.arpa"},
		{"10.100.2.255/17", "0-17.100.10.in-addr.arpa"},
		//
		{"10.100.2.255/15", "100-15.10.in-addr.arpa"},
		{"10.100.2.255/14", "100-14.10.in-addr.arpa"},
		{"10.100.2.255/13", "96-13.10.in-addr.arpa"},
		{"10.100.2.255/12", "96-12.10.in-addr.arpa"},
		{"10.100.2.255/11", "96-11.10.in-addr.arpa"},
		{"10.100.2.255/10", "64-10.10.in-addr.arpa"},
		{"10.100.2.255/9", "0-9.10.in-addr.arpa"},
	}
	for i, tst := range tests {
		t.Run(fmt.Sprintf("%d--%s", i, tst.in), func(t *testing.T) {
			d, err := ReverseDomainName(tst.in)
			if err != nil {
				t.Errorf("Should not have errored: %v", err)
			} else if d != tst.out {
				t.Errorf("Expected '%s' but got '%s'", tst.out, d)
			}
		})
	}
}

func TestReverseErrors(t *testing.T) {
	var tests = []struct {
		in string
	}{
		{"0.0.0.0/0"},
		{"2001::/0"},
		{"4.5/16"},
		{"foo.com"},
	}
	for i, tst := range tests {
		t.Run(fmt.Sprintf("%d--%s", i, tst.in), func(t *testing.T) {
			d, err := ReverseDomainName(tst.in)
			if err == nil {
				t.Errorf("Should have errored, but didn't. Got %s", d)
			}
		})
	}
}
