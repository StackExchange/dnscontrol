package models

import (
	"testing"
)

func TestToNameservers(t *testing.T) {
	nss, e := ToNameservers([]string{"example.com", "example2.tld"})
	if e != nil {
		t.Errorf("error e %v (%v)", e, nss)
	}
	if len(nss) != 2 {
		t.Errorf("error len: %v", nss)
	}
	if nss[0].Name != "example.com" {
		t.Errorf("error 0: %v", nss[0].Name)
	}
	if nss[1].Name != "example2.tld" {
		t.Errorf("error 1: %v", nss[1].Name)
	}
}

func TestToNameservers_neg(t *testing.T) {
	nss, e := ToNameservers([]string{"example.com.", "example2.tld."})
	if e == nil {
		t.Errorf("error 3 (%v)", nss)
	}
}

func TestToNameserversStripTD(t *testing.T) {
	nss, e := ToNameserversStripTD([]string{"example.com.", "example2.tld."})
	if e != nil {
		t.Errorf("error e %v (%v)", e, nss)
	}
	if len(nss) != 2 {
		t.Errorf("error len: %v", nss)
	}
	if nss[0].Name != "example.com" {
		t.Errorf("error 0: %v", nss[0].Name)
	}
	if nss[1].Name != "example2.tld" {
		t.Errorf("error 1: %v", nss[1].Name)
	}
}

func TestToNameserversStripTD_neg(t *testing.T) {
	nss, e := ToNameserversStripTD([]string{"example.com", "example2.tld"})
	if e == nil {
		t.Errorf("error e (%v)", nss)
	}
}
