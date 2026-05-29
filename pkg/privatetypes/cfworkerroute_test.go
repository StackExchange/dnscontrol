package privatetypes

import (
	"testing"

	dnsv2 "codeberg.org/miekg/dns"
)

func TestCfWorkerRoute(t *testing.T) {
	y := &CFWORKERROUTE{Hdr: dnsv2.Header{Name: "example.org.", Class: dnsv2.ClassINET}, When: "whenWhen", Then: "ThenThen"}
	//fmt.Printf("DEBUG: %v\n", dnsv2.StringToType)
	//t.Fatalf("CFWORKERROUTE string presentations should be identical: %q", y.String())
	rry, err := dnsv2.New(y.String())
	if err != nil {
		t.Fatal(err)
	}
	if rry.String() != y.String() {
		t.Fatalf("CFWORKERROUTE string presentations should be identical:\n%s\n%s", rry.String(), y.String())
	}
}
