package spf

import "testing"

func TestDnsCache(t *testing.T) {

	cache := &dnsCache{}
	cache.dnsPut("one", "txt", []string{"a", "b", "c"})
	cache.dnsPut("two", "txt", []string{"d", "e", "f"})

	a, b := cache.dnsGet("one", "txt")
	if !(b == true && len(a) == 3 && a[0] == "a" && a[1] == "b" && a[2] == "c") {
		t.Errorf("one-txt didn't work")
	}

	a, b = cache.dnsGet("two", "txt")
	if !(b == true && len(a) == 3 && a[0] == "d" && a[1] == "e" && a[2] == "f") {
		t.Errorf("one-txt didn't work")
	}

	a, b = cache.dnsGet("three", "txt")
	if !(b == false) {
		t.Errorf("three-txt didn't work")
	}

	a, b = cache.dnsGet("two", "not")
	if !(b == false) {
		t.Errorf("two-not didn't work")
	}

}
