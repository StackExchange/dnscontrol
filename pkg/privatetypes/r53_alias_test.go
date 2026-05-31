package privatetypes

import (
	"testing"

	dnsv2 "codeberg.org/miekg/dns"
)

func TestR53Alias(t *testing.T) {
	y := &R53_ALIAS{
		Hdr:              dnsv2.Header{Name: "example.org.", Class: dnsv2.ClassINET},
		AliasType:        "A",
		Target:           "kyle.example.com.",
		EvalTargetHealth: "false",
		ZoneID:           "Z1234567890",
	}
	rry, err := dnsv2.New(y.String())
	if err != nil {
		t.Fatal(err)
	}
	if rry.String() != y.String() {
		t.Fatalf("R53_ALIAS string presentations should be identical:\n%s\n%s", rry.String(), y.String())
	}
}
