package privatetypes

import (
	"testing"

	dnsv2 "codeberg.org/miekg/dns"
)

func TestAzureAlias(t *testing.T) {
	y := &AZURE_ALIAS{Hdr: dnsv2.Header{Name: "example.org.", Class: dnsv2.ClassINET}, Target: "example.com."}
	rry, err := dnsv2.New(y.String())
	if err != nil {
		t.Fatal(err)
	}
	if rry.String() != y.String() {
		t.Fatal("AZURE_ALIAS string presentations should be identical")
	}
}
