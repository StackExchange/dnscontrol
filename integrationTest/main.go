package main

import (
	"flag"
	"log"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/nameservers"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/config"
	_ "github.com/StackExchange/dnscontrol/providers/google"
	"github.com/miekg/dns/dnsutil"
)

func main() {
	jsons, err := config.LoadProviderConfigs("providers.json")
	if err != nil {
		log.Fatalf("Error loading provider configs: %s", err)
	}
	for name, cfg := range jsons {
		log.Printf("Testing %s on %s (%s)", cfg["domain"], name, cfg["providerType"])
		provider, err := providers.CreateDNSProvider(cfg["providerType"], cfg, nil)
		if err != nil {
			log.Fatal(err)
		}
		runTests(provider, cfg["domain"])
	}
}

var dual = flag.Bool("dualProviders", false, "Set true to simulate a second DNS Provider")
var thourough = flag.Bool("query", false, "Actually query dns servers to verify results")

func runTests(prv providers.DNSServiceProvider, domainName string) {
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
		log.Printf("   Running test %d: %s", i, tst.Desc)
		dom, _ := dc.Copy()
		for _, r := range tst.Records {
			rc := models.RecordConfig(*r)
			rc.NameFQDN = dnsutil.AddOrigin(rc.Name, domainName)
			dom.Records = append(dom.Records, &rc)
		}
		dom2, _ := dom.Copy()
		corrections, err := prv.GetDomainCorrections(dom)
		if err != nil {
			log.Printf("Error! %s", err)
		}
		for _, c := range corrections {
			log.Println(c.Msg)
			err = c.F()
			if err != nil {
				log.Printf("Error! %s", err)
				break
			}
		}
		//run a second time and expect zero corrections
		corrections, err = prv.GetDomainCorrections(dom2)
		if err != nil {
			log.Printf("Error! %s", err)
		}
		if len(corrections) != 0 {
			log.Printf("Expected 0 corrections on second run, but found %d.", len(corrections))
		}

	}
}

type TestCase struct {
	Desc    string
	Records []*rec
}

type rec models.RecordConfig

func a(name, target string) *rec {
	return &rec{
		Name:   name,
		Type:   "A",
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
}
