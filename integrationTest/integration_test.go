package main

import (
	"flag"
	"log"
	"testing"

	"fmt"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/nameservers"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/config"
	_ "github.com/StackExchange/dnscontrol/providers/google"
	"github.com/miekg/dns/dnsutil"
)

func TestDNSProviders(t *testing.T) {
	jsons, err := config.LoadProviderConfigs("providers.json")
	if err != nil {
		log.Fatalf("Error loading provider configs: %s", err)
	}
	for name, cfg := range jsons {
		t.Run(fmt.Sprintf("%s(%s)", name, cfg["domain"]), func(t *testing.T) {
			provider, err := providers.CreateDNSProvider(cfg["providerType"], cfg, nil)
			if err != nil {
				t.Fatal(err)
			}
			runTests(t, provider, cfg["domain"])
		})
	}
}

var dual = flag.Bool("dualProviders", false, "Set true to simulate a second DNS Provider")
var thourough = flag.Bool("query", false, "Actually query dns servers to verify results")

func runTests(t *testing.T, prv providers.DNSServiceProvider, domainName string) {
	dc := &models.DomainConfig{
		Name: domainName,
	}
	// fix up nameservers
	ns, err := prv.GetNameservers(domainName)
	if err != nil {
		log.Println("Failed getting nameservers", err)
		return
	}
	if *dual {
		ns = append(ns, models.StringsToNameservers([]string{"ns1.foo.com", "ns2.foo.org"})...)
	}
	dc.Nameservers = ns
	nameservers.AddNSRecords(dc)
	// run tests one at a time
	for i, tst := range tests {
		if t.Failed() {
			break
		}
		t.Run(fmt.Sprintf("%d: %s", i, tst.Desc), func(t *testing.T) {
			dom, _ := dc.Copy()
			for _, r := range tst.Records {
				rc := models.RecordConfig(*r)
				rc.NameFQDN = dnsutil.AddOrigin(rc.Name, domainName)
				dom.Records = append(dom.Records, &rc)
			}
			dom2, _ := dom.Copy()
			// get corrections for first time
			corrections, err := prv.GetDomainCorrections(dom)
			if err != nil {
				t.Fatal(err)
			}
			if i != 0 && len(corrections) == 0 {
				t.Fatalf("Expect changes for all tests, but got none")
			}
			for _, c := range corrections {
				err = c.F()
				if err != nil {
					t.Fatal(err)
				}
			}
			//run a second time and expect zero corrections
			corrections, err = prv.GetDomainCorrections(dom2)
			if err != nil {
				t.Fatal(err)
			}
			if len(corrections) != 0 {
				t.Fatalf("Expected 0 corrections on second run, but found %d.", len(corrections))

			}
		})

	}
}

type TestCase struct {
	Desc    string
	Records []*rec
}

type rec models.RecordConfig

func a(name, target string) *rec {
	return makeRec(name, target, "A")
}

func cname(name, target string) *rec {
	return makeRec(name, target, "CNAME")
}

func makeRec(name, target, typ string) *rec {
	return &rec{
		Name:   name,
		Type:   typ,
		Target: target,
		TTL:    300,
	}
}

func (r *rec) ttl(t uint32) *rec {
	r.TTL = t
	return r
}

func tc(desc string, recs ...*rec) *TestCase {
	return &TestCase{
		Desc:    desc,
		Records: recs,
	}
}

var tests = []*TestCase{
	// A
	tc("Empty"),
	tc("Create an A record", a("@", "1.1.1.1")),
	tc("Change it", a("@", "1.2.3.4")),
	tc("Add another", a("@", "1.2.3.4"), a("www", "1.2.3.4")),
	tc("Add another(same name)", a("@", "1.2.3.4"), a("www", "1.2.3.4"), a("www", "5.6.7.8")),
	tc("Change a ttl", a("@", "1.2.3.4").ttl(100), a("www", "1.2.3.4"), a("www", "5.6.7.8")),
	tc("Change single target from set", a("@", "1.2.3.4").ttl(100), a("www", "2.2.2.2"), a("www", "5.6.7.8")),
	tc("Change all ttls", a("@", "1.2.3.4").ttl(500), a("www", "2.2.2.2").ttl(400), a("www", "5.6.7.8").ttl(400)),
	tc("Delete one", a("@", "1.2.3.4").ttl(500), a("www", "5.6.7.8").ttl(400)),
	tc("Add back and change ttl", a("www", "5.6.7.8").ttl(700), a("www", "1.2.3.4").ttl(700)),
	tc("Change targets and ttls", a("www", "1.1.1.1"), a("www", "2.2.2.2")),
	// CNAMES
	tc("Empty"),
	tc("Create a CNAME", cname("foo", "google.com.")),
	tc("Change it", cname("foo", "google2.com.")),
	tc("Change to A record", a("foo", "1.2.3.4")),
	tc("Change back to CNAME", cname("foo", "google.com.")),
	// INTERNATIONAL: TODO: figure out how we want to present/handle these. As much as possible I want the human form in our DSLs
	// I suspect providers will vary on if they want things raw or punycoded. Don't really want each provider to have to process records, but maybe we do.
	// A helper like `domain.Records.PunyCode()` may be all we need for providers that require it.
	tc("Internationalized name", a("ööö", "1.2.3.4")),
	tc("Change IDN", a("ööö", "2.2.2.2")),
	tc("Internationalized CNAME Target", cname("a", "ööö.com.")),
	tc("IDN CNAME AND Target", cname("öoö", "ööö.ööö.")),
}
