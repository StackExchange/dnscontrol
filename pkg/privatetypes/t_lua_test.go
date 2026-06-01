package privatetypes

import (
	"testing"

	dnsv2 "codeberg.org/miekg/dns"
)

func TestLua(t *testing.T) {
	y := &LUA{Hdr: dnsv2.Header{Name: "example.org.", Class: dnsv2.ClassINET}}
	rry, err := dnsv2.New(y.String())
	if err != nil {
		t.Fatal(err)
	}
	if rry.String() != y.String() {
		t.Fatalf("LUA string presentations should be identical:\n%q\n%q", rry.String(), y.String())
	}
}
