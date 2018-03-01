package main

import (
	"flag"
	"testing"

	"fmt"

	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/nameservers"
	"github.com/StackExchange/dnscontrol/providers"
	_ "github.com/StackExchange/dnscontrol/providers/_all"
	"github.com/StackExchange/dnscontrol/providers/config"
	"github.com/miekg/dns/dnsutil"
	"github.com/pkg/errors"
)

var providerToRun = flag.String("provider", "", "Provider to run")
var startIdx = flag.Int("start", 0, "Test number to begin with")
var endIdx = flag.Int("end", 0, "Test index to stop after")
var verbose = flag.Bool("verbose", false, "Print corrections as you run them")

func init() {
	flag.Parse()
}

func getProvider(t *testing.T) (providers.DNSServiceProvider, string, map[int]bool) {
	if *providerToRun == "" {
		t.Log("No provider specified with -provider")
		return nil, "", nil
	}
	jsons, err := config.LoadProviderConfigs("providers.json")
	if err != nil {
		t.Fatalf("Error loading provider configs: %s", err)
	}
	fails := map[int]bool{}
	for name, cfg := range jsons {
		if *providerToRun != name {
			continue
		}
		provider, err := providers.CreateDNSProvider(name, cfg, nil)
		if err != nil {
			t.Fatal(err)
		}
		if f := cfg["knownFailures"]; f != "" {
			for _, s := range strings.Split(f, ",") {
				i, err := strconv.Atoi(s)
				if err != nil {
					t.Fatal(err)
				}
				fails[i] = true
			}
		}
		return provider, cfg["domain"], fails
	}
	t.Fatalf("Provider %s not found", *providerToRun)
	return nil, "", nil
}

func TestDNSProviders(t *testing.T) {
	provider, domain, fails := getProvider(t)
	if provider == nil {
		return
	}
	t.Run(fmt.Sprintf("%s", domain), func(t *testing.T) {
		runTests(t, provider, domain, fails)
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

func runTests(t *testing.T, prv providers.DNSServiceProvider, domainName string, knownFailures map[int]bool) {
	dc := getDomainConfigWithNameservers(t, prv, domainName)
	// run tests one at a time
	end := *endIdx
	tests := makeTests(t)
	if end == 0 || end >= len(tests) {
		end = len(tests) - 1
	}
	for i := *startIdx; i <= end; i++ {
		tst := tests[i]
		if t.Failed() {
			break
		}
		t.Run(fmt.Sprintf("%d: %s", i, tst.Desc), func(t *testing.T) {
			skipVal := false
			if knownFailures[i] {
				t.Log("SKIPPING VALIDATION FOR KNOWN FAILURE CASE")
				skipVal = true
			}
			dom, _ := dc.Copy()
			for _, r := range tst.Records {
				rc := models.RecordConfig(*r)
				rc.NameFQDN = dnsutil.AddOrigin(rc.Name, domainName)
				if strings.Contains(rc.Target, "**current-domain**") {
					rc.Target = strings.Replace(rc.Target, "**current-domain**", domainName, 1) + "."
				}
				dom.Records = append(dom.Records, &rc)
			}
			dom.IgnoredLabels = tst.IgnoredLabels
			models.PostProcessRecords(dom.Records)
			dom2, _ := dom.Copy()
			// get corrections for first time
			corrections, err := prv.GetDomainCorrections(dom)
			if err != nil {
				t.Fatal(errors.Wrap(err, "runTests"))
			}
			if !skipVal && i != *startIdx && len(corrections) == 0 {
				if tst.Desc != "Empty" {
					// There are "no corrections" if the last test was programatically
					// skipped.  We detect this (possibly inaccurately) by checking to
					// see if .Desc is "Empty".
					t.Fatalf("Expect changes for all tests, but got none")
				}
			}
			for _, c := range corrections {
				if *verbose {
					t.Log(c.Msg)
				}
				err = c.F()
				if !skipVal && err != nil {
					t.Fatal(err)
				}
			}
			// run a second time and expect zero corrections
			corrections, err = prv.GetDomainCorrections(dom2)
			if err != nil {
				t.Fatal(err)
			}
			if !skipVal && len(corrections) != 0 {
				t.Logf("Expected 0 corrections on second run, but found %d.", len(corrections))
				for i, c := range corrections {
					t.Logf("#%d: %s", i, c.Msg)
				}
				t.FailNow()
			}
		})
	}
}

func TestDualProviders(t *testing.T) {
	p, domain, _ := getProvider(t)
	if p == nil {
		return
	}
	dc := getDomainConfigWithNameservers(t, p, domain)
	// clear everything
	run := func() {
		dom, _ := dc.Copy()
		cs, err := p.GetDomainCorrections(dom)
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
		t.Logf("Expect no corrections on second run, but found %d.", len(cs))
		for i, c := range cs {
			t.Logf("#%d: %s", i, c.Msg)
		}
		t.FailNow()
	}
}

type TestCase struct {
	Desc          string
	Records       []*rec
	IgnoredLabels []string
}

type rec models.RecordConfig

func a(name, target string) *rec {
	return makeRec(name, target, "A")
}

func cname(name, target string) *rec {
	return makeRec(name, target, "CNAME")
}

func alias(name, target string) *rec {
	return makeRec(name, target, "ALIAS")
}

func r53alias(name, aliasType, target string) *rec {
	r := makeRec(name, target, "R53_ALIAS")
	r.R53Alias = map[string]string{
		"type": aliasType,
	}
	return r
}

func ns(name, target string) *rec {
	return makeRec(name, target, "NS")
}

func mx(name string, prio uint16, target string) *rec {
	r := makeRec(name, target, "MX")
	r.MxPreference = prio
	return r
}

func ptr(name, target string) *rec {
	return makeRec(name, target, "PTR")
}

func srv(name string, priority, weight, port uint16, target string) *rec {
	r := makeRec(name, target, "SRV")
	r.SrvPriority = priority
	r.SrvWeight = weight
	r.SrvPort = port
	return r
}

func txt(name, target string) *rec {
	// FYI: This must match the algorithm in pkg/js/helpers.js TXT.
	r := makeRec(name, target, "TXT")
	r.TxtStrings = []string{target}
	return r
}

func txtmulti(name string, target []string) *rec {
	// FYI: This must match the algorithm in pkg/js/helpers.js TXT.
	r := makeRec(name, target[0], "TXT")
	r.TxtStrings = target
	return r
}

func caa(name string, tag string, flag uint8, target string) *rec {
	r := makeRec(name, target, "CAA")
	r.CaaFlag = flag
	r.CaaTag = tag
	return r
}

func tlsa(name string, usage, selector, matchingtype uint8, target string) *rec {
	r := makeRec(name, target, "TLSA")
	r.TlsaUsage = usage
	r.TlsaSelector = selector
	r.TlsaMatchingType = matchingtype
	r.Target = target
	return r
}

func ignore(name string) *rec {
	return &rec{
		Name: name,
		Type: "IGNORE",
	}
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
	var records []*rec
	var ignored []string
	for _, r := range recs {
		if r.Type == "IGNORE" {
			ignored = append(ignored, r.Name)
		} else {
			records = append(records, r)
		}
	}
	return &TestCase{
		Desc:          desc,
		Records:       records,
		IgnoredLabels: ignored,
	}
}

func manyA(namePattern, target string, n int) []*rec {
	recs := []*rec{}
	for i := 0; i < n; i++ {
		recs = append(recs, makeRec(fmt.Sprintf(namePattern, i), target, "A"))
	}
	return recs
}

func makeTests(t *testing.T) []*TestCase {
	// ALWAYS ADD TO BOTTOM OF LIST. Order and indexes matter.
	tests := []*TestCase{
		// A
		tc("Empty"),
		tc("Create an A record", a("@", "1.1.1.1")),
		tc("Change it", a("@", "1.2.3.4")),
		tc("Add another", a("@", "1.2.3.4"), a("www", "1.2.3.4")),
		tc("Add another(same name)", a("@", "1.2.3.4"), a("www", "1.2.3.4"), a("www", "5.6.7.8")),
		tc("Change a ttl", a("@", "1.2.3.4").ttl(1000), a("www", "1.2.3.4"), a("www", "5.6.7.8")),
		tc("Change single target from set", a("@", "1.2.3.4").ttl(1000), a("www", "2.2.2.2"), a("www", "5.6.7.8")),
		tc("Change all ttls", a("@", "1.2.3.4").ttl(500), a("www", "2.2.2.2").ttl(400), a("www", "5.6.7.8").ttl(400)),
		tc("Delete one", a("@", "1.2.3.4").ttl(500), a("www", "5.6.7.8").ttl(400)),
		tc("Add back and change ttl", a("www", "5.6.7.8").ttl(700), a("www", "1.2.3.4").ttl(700)),
		tc("Change targets and ttls", a("www", "1.1.1.1"), a("www", "2.2.2.2")),
		tc("Create wildcard", a("*", "1.2.3.4"), a("www", "1.1.1.1")),
		tc("Delete wildcard", a("www", "1.1.1.1")),

		// CNAMES
		tc("Empty"),
		tc("Create a CNAME", cname("foo", "google.com.")),
		tc("Change it", cname("foo", "google2.com.")),
		tc("Change to A record", a("foo", "1.2.3.4")),
		tc("Change back to CNAME", cname("foo", "google.com.")),
		tc("Record pointing to @", cname("foo", "**current-domain**")),

		// NS
		tc("Empty"),
		tc("NS for subdomain", ns("xyz", "ns2.foo.com.")),
		tc("Dual NS for subdomain", ns("xyz", "ns2.foo.com."), ns("xyz", "ns1.foo.com.")),
		tc("NS Record pointing to @", ns("foo", "**current-domain**")),

		// IDNAs
		tc("Empty"),
		tc("Internationalized name", a("ööö", "1.2.3.4")),
		tc("Change IDN", a("ööö", "2.2.2.2")),
		tc("Internationalized CNAME Target", cname("a", "ööö.com.")),
		tc("IDN CNAME AND Target", cname("öoö", "ööö.企业.")),

		// MX
		tc("Empty"),
		tc("MX record", mx("@", 5, "foo.com.")),
		tc("Second MX record, same prio", mx("@", 5, "foo.com."), mx("@", 5, "foo2.com.")),
		tc("3 MX", mx("@", 5, "foo.com."), mx("@", 5, "foo2.com."), mx("@", 15, "foo3.com.")),
		tc("Delete one", mx("@", 5, "foo2.com."), mx("@", 15, "foo3.com.")),
		tc("Change to other name", mx("@", 5, "foo2.com."), mx("mail", 15, "foo3.com.")),
		tc("Change Preference", mx("@", 7, "foo2.com."), mx("mail", 15, "foo3.com.")),
		tc("Record pointing to @", mx("foo", 8, "**current-domain**")),
	}

	// PTR
	if !providers.ProviderHasCabability(*providerToRun, providers.CanUsePTR) {
		t.Log("Skipping PTR Tests because provider does not support them")
	} else {
		tests = append(tests, tc("Empty"),
			tc("Create PTR record", ptr("4", "foo.com.")),
			tc("Modify PTR record", ptr("4", "bar.com.")),
		)
	}

	// ALIAS
	if !providers.ProviderHasCabability(*providerToRun, providers.CanUseAlias) {
		t.Log("Skipping ALIAS Tests because provider does not support them")
	} else {
		tests = append(tests, tc("Empty"),
			tc("ALIAS at root", alias("@", "foo.com.")),
			tc("change it", alias("@", "foo2.com.")),
			tc("ALIAS at subdomain", alias("test", "foo.com.")),
		)
	}

	// SRV
	if !providers.ProviderHasCabability(*providerToRun, providers.CanUseSRV) {
		t.Log("Skipping SRV Tests because provider does not support them")
	} else {
		tests = append(tests, tc("Empty"),
			tc("SRV record", srv("_sip._tcp", 5, 6, 7, "foo.com.")),
			tc("Second SRV record, same prio", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 5, 60, 70, "foo2.com.")),
			tc("3 SRV", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 5, 60, 70, "foo2.com."), srv("_sip._tcp", 15, 65, 75, "foo3.com.")),
			tc("Delete one", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo3.com.")),
			tc("Change Target", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
			tc("Change Priority", srv("_sip._tcp", 52, 6, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
			tc("Change Weight", srv("_sip._tcp", 52, 62, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
			tc("Change Port", srv("_sip._tcp", 52, 62, 72, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
		)
	}

	// CAA
	if !providers.ProviderHasCabability(*providerToRun, providers.CanUseCAA) {
		t.Log("Skipping CAA Tests because provider does not support them")
	} else {
		tests = append(tests, tc("Empty"),
			tc("CAA record", caa("@", "issue", 0, "letsencrypt.org")),
			tc("CAA change tag", caa("@", "issuewild", 0, "letsencrypt.org")),
			tc("CAA change target", caa("@", "issuewild", 0, "example.com")),
			tc("CAA change flag", caa("@", "issuewild", 128, "example.com")),
			tc("CAA many records", caa("@", "issue", 0, "letsencrypt.org"), caa("@", "issuewild", 0, ";"), caa("@", "iodef", 128, "mailto:test@example.com")),
			tc("CAA delete", caa("@", "issue", 0, "letsencrypt.org")),
		)
	}

	// TLSA
	if !providers.ProviderHasCabability(*providerToRun, providers.CanUseTLSA) {
		t.Log("Skipping TLSA Tests because provider does not support them")
	} else {
		sha256hash := strings.Repeat("0123456789abcdef", 4)
		sha512hash := strings.Repeat("0123456789abcdef", 8)
		reversedSha512 := strings.Repeat("fedcba9876543210", 8)
		tests = append(tests, tc("Empty"),
			tc("TLSA record", tlsa("_443._tcp", 3, 1, 1, sha256hash)),
			tc("TLSA change usage", tlsa("_443._tcp", 2, 1, 1, sha256hash)),
			tc("TLSA change selector", tlsa("_443._tcp", 2, 0, 1, sha256hash)),
			tc("TLSA change matchingtype", tlsa("_443._tcp", 2, 0, 2, sha512hash)),
			tc("TLSA change certificate", tlsa("_443._tcp", 2, 0, 2, reversedSha512)),
		)
	}

	// Case
	tests = append(tests, tc("Empty"),
		tc("Empty"),
		tc("Create CAPS", mx("BAR", 5, "BAR.com.")),
		tc("Downcase label", mx("bar", 5, "BAR.com."), a("decoy", "1.1.1.1")),
		tc("Downcase target", mx("bar", 5, "bar.com."), a("decoy", "2.2.2.2")),
		tc("Upcase both", mx("BAR", 5, "BAR.COM."), a("decoy", "3.3.3.3")),
		// The decoys are required so that there is at least one actual change in each tc.
	)

	// Test large zonefiles.
	// Mostly to test paging. Many providers page at 100
	// Known page sizes:
	//  - gandi: 100
	skip := map[string]bool{
		"NS1": true, // ns1 free acct only allows 50 records
	}
	if skip[*providerToRun] {
		t.Log("Skipping Large record count Tests because provider does not support them")
	} else {
		tests = append(tests, tc("Empty"),
			tc("99 records", manyA("rec%04d", "1.2.3.4", 99)...),
			tc("100 records", manyA("rec%04d", "1.2.3.4", 100)...),
			tc("101 records", manyA("rec%04d", "1.2.3.4", 101)...),
		)
	}

	// NB(tlim): To temporarily skip most of the tests, insert a line like this:
	//tests = nil

	// TXT (single)
	tests = append(tests, tc("Empty"),
		tc("Empty"),
		tc("Create a TXT", txt("foo", "simple")),
		tc("Change a TXT", txt("foo", "changed")),
		tc("Empty"),
		tc("Create a TXT with spaces", txt("foo", "with spaces")),
		tc("Change a TXT with spaces", txt("foo", "with whitespace")),
		tc("Create 1 TXT as array", txtmulti("foo", []string{"simple"})),
	)

	// TXTMulti
	if !providers.ProviderHasCabability(*providerToRun, providers.CanUseTXTMulti) {
		t.Log("Skipping TXTMulti Tests because provider does not support them")
	} else {
		tests = append(tests,
			tc("Empty"),
			tc("Create TXTMulti 1",
				txtmulti("foo1", []string{"simple"}),
			),
			tc("Create TXTMulti 2",
				txtmulti("foo1", []string{"simple"}),
				txtmulti("foo2", []string{"one", "two"}),
			),
			tc("Create TXTMulti 3",
				txtmulti("foo1", []string{"simple"}),
				txtmulti("foo2", []string{"one", "two"}),
				txtmulti("foo3", []string{"eh", "bee", "cee"}),
			),
			tc("Create TXTMulti with quotes",
				txtmulti("foo1", []string{"simple"}),
				txtmulti("foo2", []string{"o\"ne", "tw\"o"}),
				txtmulti("foo3", []string{"eh", "bee", "cee"}),
			),
			tc("Change TXTMulti",
				txtmulti("foo1", []string{"dimple"}),
				txtmulti("foo2", []string{"fun", "two"}),
				txtmulti("foo3", []string{"eh", "bzz", "cee"}),
			),
			tc("Empty"),
		)
	}

	// ignored records
	tests = append(tests,
		tc("Empty"),
		tc("Create some records", txt("foo", "simple"), a("foo", "1.2.3.4")),
		tc("Add a new record - ignoring foo", a("bar", "1.2.3.4"), ignore("foo")),
	)

	// R53_ALIAS
	if !providers.ProviderHasCabability(*providerToRun, providers.CanUseRoute53Alias) {
		t.Log("Skipping Route53 ALIAS Tests because provider does not support them")
	} else {
		tests = append(tests, tc("Empty"),
			tc("create dependent records", a("foo", "1.2.3.4"), a("quux", "2.3.4.5")),
			tc("ALIAS to A record in same zone", a("foo", "1.2.3.4"), a("quux", "2.3.4.5"), r53alias("bar", "A", "foo.**current-domain**")),
			tc("change it", a("foo", "1.2.3.4"), a("quux", "2.3.4.5"), r53alias("bar", "A", "quux.**current-domain**")),
		)
	}

	return tests
}
