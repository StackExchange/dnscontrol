package privatetypes

import (
	"testing"

	dnsv2 "codeberg.org/miekg/dns"
)

func TestAlias(t *testing.T) {
	y := &ALIAS{Hdr: dnsv2.Header{Name: "example.org.", Class: dnsv2.ClassINET}, Target: "example.com."}
	rry, err := dnsv2.New(y.String())
	if err != nil {
		t.Fatal(err)
	}
	if rry.String() != y.String() {
		t.Fatalf("ALIAS string presentations should be identical: %q %q", rry.String(), y.String())
	}
}
