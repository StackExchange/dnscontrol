package privatetypes

import (
	"testing"

	dnsv2 "codeberg.org/miekg/dns"
)

func TestCfWorkerRoute(t *testing.T) {
	y := &CFWORKERROUTE{Hdr: dnsv2.Header{Name: "example.org.", Class: dnsv2.ClassINET}, When: "whenWhen", Then: "ThenThen"}
	rry, err := dnsv2.New(y.String())
	if err != nil {
		t.Fatal(err)
	}
	if rry.String() != y.String() {
		t.Fatalf("CFWORKERROUTE string presentations should be identical: %q %q", rry.String(), y.String())
	}
}
