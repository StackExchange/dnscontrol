package privatetypes

import (
	"testing"

	dnsv2 "codeberg.org/miekg/dns"
)

func TestBunny_DNS_Pz(t *testing.T) {
	y := &BUNNY_DNS_PZ{Hdr: dnsv2.Header{Name: "example.org.", Class: dnsv2.ClassINET}}
	rry, err := dnsv2.New(y.String())
	if err != nil {
		t.Fatal(err)
	}
	if rry.String() != y.String() {
		t.Fatalf("BUNNY_DNS_PZ string presentations should be identical:\n%q\n%q", rry.String(), y.String())
	}
}
