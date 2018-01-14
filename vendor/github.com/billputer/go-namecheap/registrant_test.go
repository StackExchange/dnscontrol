package namecheap

import (
	"net/url"
	"testing"
)

func TestAddValues(t *testing.T) {
	reg := newRegistrant(
		"r", "m",
		"10 Park Ave.",
		"Apt. 3F",
		"NY", "New York", "10001", "US",
		"9125357070", "joe.dirt1@gmail.com",
	)

	u := url.Values{}
	reg.addValues(u)

	if a, n := u.Get("RegistrantFirstName"), "r"; a != n {
		t.Errorf("expected %s, got %s", n, a)
	}

	if a, n := u.Get("TechFirstName"), "r"; a != n {
		t.Errorf("expected %s, got %s", n, a)
	}

	if a, n := u.Get("AdminFirstName"), "r"; a != n {
		t.Errorf("expected %s, got %s", n, a)
	}

	if a, n := u.Get("AuxBillingFirstName"), "r"; a != n {
		t.Errorf("expected %s, got %s", n, a)
	}

	reg = new(Registrant)
	u = url.Values{}

	if err := reg.addValues(u); err == nil {
		t.Error("Should have returned error. All fields empty")
	}
}
