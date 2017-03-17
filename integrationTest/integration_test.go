package main

import (
	"flag"
	"testing"

	"fmt"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/nameservers"
	"github.com/StackExchange/dnscontrol/providers"
	_ "github.com/StackExchange/dnscontrol/providers/_all"
	"github.com/StackExchange/dnscontrol/providers/config"
	"github.com/miekg/dns/dnsutil"
)

var providerToRun = flag.String("provider", "", "Provider to run")

func init() {
	flag.Parse()
}

func getProvider(t *testing.T) (providers.DNSServiceProvider, string) {
	if *providerToRun == "" {
		t.Log("No provider specified with -provider")
		return nil, ""
	}
	jsons, err := config.LoadProviderConfigs("providers.json")
	if err != nil {
		t.Fatalf("Error loading provider configs: %s", err)
	}
	for name, cfg := range jsons {
		if *providerToRun != name {
			continue
		}
		provider, err := providers.CreateDNSProvider(name, cfg, nil)
		if err != nil {
			t.Fatal(err)
		}
		return provider, cfg["domain"]
	}
	t.Fatalf("Provider %s not found", *providerToRun)
	return nil, ""
}

func TestDNSProviders(t *testing.T) {
	provider, domain := getProvider(t)
	if provider == nil {
		return
	}
	t.Run(fmt.Sprintf("%s", domain), func(t *testing.T) {
		runTests(t, provider, domain)
	})

}

func getDomainConfigWithNameservers(t *testing.T, prv providers.DNSServiceProvider, domainName string) *models.DomainConfig {
	dc := &models.DomainConfig{
		Name: domainName,
	}
	// fix up nameservers
	ns, err := prv.GetNameservers(domainName)
	if err != nil {
		t.Fatal("Failed getting nameservers", err)
	}
	dc.Nameservers = ns
	nameservers.AddNSRecords(dc)
	return dc
}

func runTests(t *testing.T, prv providers.DNSServiceProvider, domainName string) {
	dc := getDomainConfigWithNameservers(t, prv, domainName)
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

func TestDualProviders(t *testing.T) {
	p, domain := getProvider(t)
	if p == nil {
		return
	}
	dc := getDomainConfigWithNameservers(t, p, domain)
	// clear everything
	run := func() {
		cs, err := p.GetDomainCorrections(dc)
		if err != nil {
			t.Fatal(err)
		}
		for i, c := range cs {
			t.Logf("#%d: %s", i+1, c.Msg)
			if err = c.F(); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Log("Clearing everything")
	run()
	// add bogus nameservers
	dc.Records = []*models.RecordConfig{}
	dc.Nameservers = append(dc.Nameservers, models.StringsToNameservers([]string{"ns1.otherdomain.tld", "ns2.otherdomain.tld"})...)
	nameservers.AddNSRecords(dc)
	t.Log("Adding nameservers from another provider")
	run()
	// run again to make sure no corrections
	t.Log("Running again to ensure stability")
	cs, err := p.GetDomainCorrections(dc)
	if err != nil {
		t.Fatal(err)
	}
	if len(cs) != 0 {
		t.Fatal("Expect no corrections on second run")
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

	//IDNAs
	tc("Internationalized name", a("ööö", "1.2.3.4")),
	tc("Change IDN", a("ööö", "2.2.2.2")),
	tc("Internationalized CNAME Target", cname("a", "ööö.com.")),
	tc("IDN CNAME AND Target", cname("öoö", "ööö.ööö.")),
}
