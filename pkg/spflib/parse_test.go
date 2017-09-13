package spflib

import (
	"strings"
	"testing"

	"github.com/StackExchange/dnscontrol/pkg/dnsresolver"
)

func TestParse(t *testing.T) {
	dnsres, err := dnsresolver.NewResolverPreloaded("testdata-dns1.json")
	if err != nil {
		t.Fatal(err)
	}
	rec, err := Parse(strings.Join([]string{"v=spf1",
		"ip4:198.252.206.0/24",
		"ip4:192.111.0.0/24",
		"include:_spf.google.com",
		"include:mailgun.org",
		"include:spf-basic.fogcreek.com",
		"include:mail.zendesk.com",
		"include:servers.mcsv.net",
		"include:sendgrid.net",
		"include:spf.mtasv.net",
		"~all"}, " "), dnsres)
	if err != nil {
		t.Fatal(err)
	}
	DumpSPF(rec, "")
}
