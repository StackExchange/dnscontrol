package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/miekg/dns/dnsutil"

	"github.com/StackExchange/dnscontrol/v2/models"
	"github.com/StackExchange/dnscontrol/v2/pkg/nameservers"
	"github.com/StackExchange/dnscontrol/v2/providers"
	_ "github.com/StackExchange/dnscontrol/v2/providers/_all"
	"github.com/StackExchange/dnscontrol/v2/providers/config"
)

var providerToRun = flag.String("provider", "", "Provider to run")
var startIdx = flag.Int("start", 0, "Test number to begin with")
var endIdx = flag.Int("end", 0, "Test index to stop after")
var verbose = flag.Bool("verbose", false, "Print corrections as you run them")

func init() {
	testing.Init()
	flag.Parse()
}

func getProvider(t *testing.T) (providers.DNSServiceProvider, string, map[int]bool, map[string]string) {
	if *providerToRun == "" {
		t.Log("No provider specified with -provider")
		return nil, "", nil, nil
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
		return provider, cfg["domain"], fails, cfg
	}
	t.Fatalf("Provider %s not found", *providerToRun)
	return nil, "", nil, nil
}

func TestDNSProviders(t *testing.T) {
	provider, domain, fails, cfg := getProvider(t)
	if provider == nil {
		return
	}
	t.Run(fmt.Sprintf("%s", domain), func(t *testing.T) {
		runTests(t, provider, domain, fails, cfg)
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

// testPermitted returns nil if the test is permitted, otherwise an
// error explaining why it is not.
func testPermitted(t *testing.T, p string, f TestCase) error {

	// not() and only() can't be mixed.
	if len(f.only) != 0 && len(f.not) != 0 {
		return fmt.Errorf("invalid filter: can't mix not() and only()")
	}
	// TODO(tlim): Have a separate validation pass so that such mistakes
	// are more visible?

	// If there are any required capabilities, make sure they all exist.
	if len(f.required) != 0 {
		for _, c := range f.required {
			if !providers.ProviderHasCapability(*providerToRun, c) {
				return fmt.Errorf("%s not supported", c)
			}
		}
	}

	// If there are any "only" items, you must be one of them.
	if len(f.only) != 0 {
		for _, provider := range f.only {
			if p == provider {
				return nil
			}
		}
		return fmt.Errorf("disabled by only()")
	}

	// If there are any "not" items, you must NOT be one of them.
	if len(f.not) != 0 {
		for _, provider := range f.not {
			if p == provider {
				return fmt.Errorf("excluded by not(\"%s\")", provider)
			}
		}
		return nil
	}

	return nil
}

func makeClearFilter() *TestCase {
	tc := tc("Empty")
	tc.ChangeFilter = true
	return tc
}

func runTests(t *testing.T, prv providers.DNSServiceProvider, domainName string, knownFailures map[int]bool, origConfig map[string]string) {
	dc := getDomainConfigWithNameservers(t, prv, domainName)
	// run tests one at a time
	end := *endIdx
	tests := makeTests(t)
	if end == 0 || end >= len(tests) {
		end = len(tests) - 1
	}

	curFilter := makeClearFilter()

	for i := *startIdx; i <= end; i++ {
		tst := tests[i]

		if t.Failed() { // Did the previous test fail? Stop.
			break
		}

		if tst.ChangeFilter {
			curFilter = tst
			// Reset the filter. Keep going, to execute the "Empty".
		}

		skipVal := false // Skip validation

		if err := testPermitted(t, *providerToRun, *curFilter); err != nil {
			t.Logf("%s%s",
				strings.ReplaceAll(fmt.Sprintf("%d: %s:", i, tst.Desc), " ", "_"),
				fmt.Sprintf(" **** SKIPPING: %v", err),
			)
			// We skip by removing the records. As a result, this test
			// becomes the same as "Empty", which does not require certain
			// validations (that the test MUST have at least one change,
			// that the re-test MUST NOT have at least one change).
			tst.Records = nil
			skipVal = true
		}

		t.Run(fmt.Sprintf("%d: %s", i, tst.Desc), func(t *testing.T) {
			dom, _ := dc.Copy()
			for _, r := range tst.Records {
				rc := models.RecordConfig(*r)
				if strings.Contains(rc.GetTargetField(), "**current-domain**") {
					_ = rc.SetTarget(strings.Replace(rc.GetTargetField(), "**current-domain**", domainName, 1) + ".")
				}
				if strings.Contains(rc.GetTargetField(), "**current-domain-no-trailing**") {
					_ = rc.SetTarget(strings.Replace(rc.GetTargetField(), "**current-domain-no-trailing**", domainName, 1))
				}
				if strings.Contains(rc.GetLabelFQDN(), "**current-domain**") {
					rc.SetLabelFromFQDN(strings.Replace(rc.GetLabelFQDN(), "**current-domain**", domainName, 1), domainName)
				}
				if providers.ProviderHasCapability(*providerToRun, providers.CanUseAzureAlias) {
					if strings.Contains(rc.GetTargetField(), "**subscription-id**") {
						_ = rc.SetTarget(strings.Replace(rc.GetTargetField(), "**subscription-id**", origConfig["SubscriptionID"], 1))
					}
					if strings.Contains(rc.GetTargetField(), "**resource-group**") {
						_ = rc.SetTarget(strings.Replace(rc.GetTargetField(), "**resource-group**", origConfig["ResourceGroup"], 1))
					}
				}
				dom.Records = append(dom.Records, &rc)
			}
			dom.IgnoredLabels = tst.IgnoredLabels
			models.PostProcessRecords(dom.Records)
			dom2, _ := dom.Copy()
			// get corrections for first time
			corrections, err := prv.GetDomainCorrections(dom)
			if err != nil {
				t.Fatal(fmt.Errorf("runTests: %w", err))
			}
			if !skipVal && (i != *startIdx && len(corrections) == 0) {
				if tst.Desc != "Empty" {
					// There are "no corrections" if the last test was programmatically
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
			if len(corrections) != 0 {
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
	p, domain, _, _ := getProvider(t)
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
	dc.Nameservers = append(dc.Nameservers, models.StringsToNameservers([]string{"ns1.example.com", "ns2.example.com"})...)
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
	ChangeFilter bool // If true, reset the filter before doing the test.
	// TEST:
	Desc          string
	Records       []*rec
	IgnoredLabels []string
	// FILTER:
	required []providers.Capability
	only     []string
	not      []string
}

type rec models.RecordConfig

func (r *rec) GetLabel() string {
	return r.Name
}

func (r *rec) SetLabel(label, domain string) {
	r.Name = label
	r.NameFQDN = dnsutil.AddOrigin(label, "**current-domain**")
}

func (r *rec) SetTarget(target string) {
	r.Target = target
}

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

func azureAlias(name, aliasType, target string) *rec {
	r := makeRec(name, target, "AZURE_ALIAS")
	r.AzureAlias = map[string]string{
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

func naptr(name string, order uint16, preference uint16, flags string, service string, regexp string, target string) *rec {
	r := makeRec(name, target, "NAPTR")
	r.NaptrOrder = order
	r.NaptrPreference = preference
	r.NaptrFlags = flags
	r.NaptrService = service
	r.NaptrRegexp = regexp
	return r
}

func srv(name string, priority, weight, port uint16, target string) *rec {
	r := makeRec(name, target, "SRV")
	r.SrvPriority = priority
	r.SrvWeight = weight
	r.SrvPort = port
	return r
}

func sshfp(name string, algorithm uint8, fingerprint uint8, target string) *rec {
	r := makeRec(name, target, "SSHFP")
	r.SshfpAlgorithm = algorithm
	r.SshfpFingerprint = fingerprint
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
	return r
}

func ignore(name string) *rec {
	r := &rec{
		Type: "IGNORE",
	}
	r.SetLabel(name, "**current-domain**")
	return r
}

func makeRec(name, target, typ string) *rec {
	r := &rec{
		Type: typ,
		TTL:  300,
	}
	r.SetLabel(name, "**current-domain**")
	r.SetTarget(target)
	return r
}

func (r *rec) ttl(t uint32) *rec {
	r.TTL = t
	return r
}

func manyA(namePattern, target string, n int) []*rec {
	recs := []*rec{}
	for i := 0; i < n; i++ {
		recs = append(recs, makeRec(fmt.Sprintf(namePattern, i), target, "A"))
	}
	return recs
}

func tc(desc string, recs ...*rec) *TestCase {
	var records []*rec
	var ignored []string
	for _, r := range recs {
		if r.Type == "IGNORE" {
			ignored = append(ignored, r.GetLabel())
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

func reset(items ...interface{}) *TestCase {
	tc := makeClearFilter()
	for _, item := range items {
		switch v := item.(type) {
		case requiresFilter:
			tc.required = append(tc.required, v.cap)
		case notFilter:
			tc.not = append(tc.not, v.name)
		case onlyFilter:
			tc.only = append(tc.only, v.name)
		default:
			fmt.Printf("I don't know about type %T (%v)\n", v, v)
		}
	}
	return tc
}

type requiresFilter struct {
	cap providers.Capability
}

func requires(c providers.Capability) requiresFilter {
	return requiresFilter{cap: c}
}

type notFilter struct {
	name string
}

func not(n string) notFilter {
	return notFilter{name: n}
}

type onlyFilter struct {
	name string
}

func only(n string) onlyFilter {
	return onlyFilter{name: n}
}

//

func makeTests(t *testing.T) []*TestCase {

	sha256hash := strings.Repeat("0123456789abcdef", 4)
	sha512hash := strings.Repeat("0123456789abcdef", 8)
	reversedSha512 := strings.Repeat("fedcba9876543210", 8)

	// Each group of tests begins with reset(). It empties out the zone
	// (deletes all records) and resets the filter.

	// Start a group of tests that apply to all providers:
	//      reset()
	// Only apply to	providers that CanUseAlias.
	//      reset(requires(providers.CanUseAlias)),
	// Only apply to providers listed.
	//      reset(only("ROUTE53")),
	// Only apply to providers listed.
	//     reset(only("ROUTE53"), only("GCLOUD")),
	// Apply to all providers except ROUTE53
	//     reset(not("ROUTE53")),
	// Apply to all providers except ROUTE53 and GCLOUD
	//     reset(not("ROUTE53"), not("GCLOUD")),

	// DETAILS:
	// You can't mix not() and only()
	//     reset(not("ROUTE53"), only("GCLOUD")),  // ERROR!

	tests := []*TestCase{

		//
		// Basic functionality (add/rename/delete)
		//

		// A
		reset(),
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
		reset(),
		tc("Create a CNAME", cname("foo", "google.com.")),
		tc("Change it", cname("foo", "google2.com.")),
		tc("Change to A record", a("foo", "1.2.3.4")),
		tc("Change back to CNAME", cname("foo", "google.com.")),
		tc("Record pointing to @", cname("foo", "**current-domain**")),

		// MX
		reset(not("ACTIVEDIRECTORY_PS")),
		tc("MX record", mx("@", 5, "foo.com.")),
		tc("Second MX record, same prio", mx("@", 5, "foo.com."), mx("@", 5, "foo2.com.")),
		tc("3 MX", mx("@", 5, "foo.com."), mx("@", 5, "foo2.com."), mx("@", 15, "foo3.com.")),
		tc("Delete one", mx("@", 5, "foo2.com."), mx("@", 15, "foo3.com.")),
		tc("Change to other name", mx("@", 5, "foo2.com."), mx("mail", 15, "foo3.com.")),
		tc("Change Preference", mx("@", 7, "foo2.com."), mx("mail", 15, "foo3.com.")),
		tc("Record pointing to @", mx("foo", 8, "**current-domain**")),

		// NS
		reset(not("DNSIMPLE"), not("EXOSCALE")),
		// DNSIMPLE: Does not support NS records nor subdomains.
		tc("NS for subdomain", ns("xyz", "ns2.foo.com.")),
		tc("Dual NS for subdomain", ns("xyz", "ns2.foo.com."), ns("xyz", "ns1.foo.com.")),
		tc("NS Record pointing to @", ns("foo", "**current-domain**")),
		// ignored records
		reset(),
		tc("Create some records", txt("foo", "simple"), a("foo", "1.2.3.4")),
		tc("Add a new record - ignoring foo", a("bar", "1.2.3.4"), ignore("foo")),
		reset(),
		tc("Create some records", txt("bar.foo", "simple"), a("bar.foo", "1.2.3.4")),
		tc("Add a new record - ignoring *.foo", a("bar", "1.2.3.4"), ignore("*.foo")),

		// TXT (single)
		reset(),
		tc("Create a TXT", txt("foo", "simple")),
		tc("Change a TXT", txt("foo", "changed")),
		reset(),
		tc("Create a TXT with spaces", txt("foo", "with spaces")),
		tc("Change a TXT with spaces", txt("foo", "with whitespace")),
		tc("Create 1 TXT as array", txtmulti("foo", []string{"simple"})),
		reset(),
		tc("Create a 255-byte TXT", txt("foo", strings.Repeat("A", 255))),

		// TXT (empty)
		reset(not("DNSIMPLE"), not("CLOUDFLAREAPI")),
		tc("TXT with empty str", txt("foo1", "")),
		// https://github.com/StackExchange/dnscontrol/issues/598
		// We decided that handling an empty TXT string is not a
		// requirement. In the future we might make it a "capability" to
		// indicate which vendors fully support RFC 1035, which requires
		// that a TXT string can be empty.

		//
		// Tests that exercise the API protocol and/or code
		//

		// Case
		// The decoys are required so that there is at least one actual change in each tc.
		reset(),
		tc("Create CAPS", mx("BAR", 5, "BAR.com.")),
		tc("Downcase label", mx("bar", 5, "BAR.com."), a("decoy", "1.1.1.1")),
		tc("Downcase target", mx("bar", 5, "bar.com."), a("decoy", "2.2.2.2")),
		tc("Upcase both", mx("BAR", 5, "BAR.COM."), a("decoy", "3.3.3.3")),

		// IDNAs
		reset(not("SOFTLAYER")),
		// SOFTLAYER: fails at direct internationalization, punycode works.
		tc("Internationalized name", a("ööö", "1.2.3.4")),
		tc("Change IDN", a("ööö", "2.2.2.2")),
		tc("Internationalized CNAME Target", cname("a", "ööö.com.")),
		// IDNAs in CNAME targets
		reset(not("LINODE")),
		// LINODE: hostname validation does not allow the target domain TLD
		tc("IDN CNAME AND Target", cname("öoö", "ööö.企业.")),

		// Tests the paging code of providers.  Many providers page at 100.
		// Notes:
		//  - gandi: page size is 100, therefore we test with 99, 100, and 101
		//  - ns1: free acct only allows 50 records
		reset(not("NS1")),
		tc("99 records", manyA("rec%04d", "1.2.3.4", 99)...),
		tc("100 records", manyA("rec%04d", "1.2.3.4", 100)...),
		tc("101 records", manyA("rec%04d", "1.2.3.4", 101)...),

		// Tests for bugs in handling VERY large updates
		reset(only("ROUTE53")),
		tc("600 records", manyA("rec%04d", "1.2.3.4", 600)...),
		tc("Update 600 records", manyA("rec%04d", "1.2.3.5", 600)...),
		tc("Empty"), // Delete them all
		tc("1200 records", manyA("rec%04d", "1.2.3.4", 1200)...),
		tc("Update 1200 records", manyA("rec%04d", "1.2.3.5", 1200)...),

		//
		// CanUse* types:
		//

		// CAA
		reset(requires(providers.CanUseCAA)),
		tc("CAA record", caa("@", "issue", 0, "letsencrypt.org")),
		tc("CAA change tag", caa("@", "issuewild", 0, "letsencrypt.org")),
		tc("CAA change target", caa("@", "issuewild", 0, "example.com")),
		tc("CAA change flag", caa("@", "issuewild", 128, "example.com")),
		tc("CAA many records",
			caa("@", "issue", 0, "letsencrypt.org"),
			caa("@", "issuewild", 0, "comodoca.com"),
			caa("@", "iodef", 128, "mailto:test@example.com")),
		tc("CAA delete", caa("@", "issue", 0, "letsencrypt.org")),
		// Test support of ";" as a value
		reset(requires(providers.CanUseCAA), not("DIGITALOCEAN")),
		tc("CAA many records", caa("@", "issuewild", 0, ";")),

		// NAPTR
		reset(requires(providers.CanUseNAPTR)),
		tc("NAPTR record", naptr("test", 100, 10, "U", "E2U+sip", "!^.*$!sip:customer-service@example.com!", "example.foo.com.")),
		tc("NAPTR second record", naptr("test", 102, 10, "U", "E2U+email", "!^.*$!mailto:information@example.com!", "example.foo.com.")),
		tc("NAPTR delete record", naptr("test", 100, 10, "U", "E2U+email", "!^.*$!mailto:information@example.com!", "example.foo.com.")),
		tc("NAPTR change target", naptr("test", 100, 10, "U", "E2U+email", "!^.*$!mailto:information@example.com!", "example2.foo.com.")),
		tc("NAPTR change order", naptr("test", 103, 10, "U", "E2U+email", "!^.*$!mailto:information@example.com!", "example2.foo.com.")),
		tc("NAPTR change preference", naptr("test", 103, 20, "U", "E2U+email", "!^.*$!mailto:information@example.com!", "example2.foo.com.")),
		tc("NAPTR change flags", naptr("test", 103, 20, "A", "E2U+email", "!^.*$!mailto:information@example.com!", "example2.foo.com.")),
		tc("NAPTR change service", naptr("test", 103, 20, "A", "E2U+sip", "!^.*$!mailto:information@example.com!", "example2.foo.com.")),
		tc("NAPTR change regexp", naptr("test", 103, 20, "A", "E2U+sip", "!^.*$!sip:customer-service@example.com!", "example2.foo.com.")),

		// PTR
		reset(requires(providers.CanUsePTR), not("ACTIVEDIRECTORY_PS")),
		tc("Create PTR record", ptr("4", "foo.com.")),
		tc("Modify PTR record", ptr("4", "bar.com.")),

		// SRV
		reset(requires(providers.CanUseSRV), not("ACTIVEDIRECTORY_PS"), not("CLOUDNS")),
		tc("SRV record", srv("_sip._tcp", 5, 6, 7, "foo.com.")),
		tc("Second SRV record, same prio", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 5, 60, 70, "foo2.com.")),
		tc("3 SRV", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 5, 60, 70, "foo2.com."), srv("_sip._tcp", 15, 65, 75, "foo3.com.")),
		tc("Delete one", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo3.com.")),
		tc("Change Target", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
		tc("Change Priority", srv("_sip._tcp", 52, 6, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
		tc("Change Weight", srv("_sip._tcp", 52, 62, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
		tc("Change Port", srv("_sip._tcp", 52, 62, 72, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
		// w/ null target
		reset(not("NAMEDOTCOM"), not("HEXONET"), not("EXOSCALE")),
		tc("Null Target", srv("_sip._tcp", 52, 62, 72, "foo.com."), srv("_sip._tcp", 15, 65, 75, ".")),

		// SSHFP
		reset(requires(providers.CanUseSSHFP)),
		tc("SSHFP record",
			sshfp("@", 1, 1, "66c7d5540b7d75a1fb4c84febfa178ad99bdd67c")),
		tc("SSHFP change algorithm",
			sshfp("@", 2, 1, "66c7d5540b7d75a1fb4c84febfa178ad99bdd67c")),
		tc("SSHFP change fingerprint and type",
			sshfp("@", 2, 2, "745a635bc46a397a5c4f21d437483005bcc40d7511ff15fbfafe913a081559bc")),
		tc("SSHFP Delete one"),
		tc("SSHFP add many records",
			sshfp("@", 1, 1, "66666666666d75a1fb4c84febfa178ad99bdd67c"),
			sshfp("@", 1, 2, "777777777777797a5c4f21d437483005bcc40d7511ff15fbfafe913a081559bc"),
			sshfp("@", 2, 1, "8888888888888888fb4c84febfa178ad99bdd67c")),
		tc("SSHFP delete two",
			sshfp("@", 1, 1, "66666666666d75a1fb4c84febfa178ad99bdd67c")),

		// TLSA
		reset(requires(providers.CanUseTLSA)),
		tc("TLSA record", tlsa("_443._tcp", 3, 1, 1, sha256hash)),
		tc("TLSA change usage", tlsa("_443._tcp", 2, 1, 1, sha256hash)),
		tc("TLSA change selector", tlsa("_443._tcp", 2, 0, 1, sha256hash)),
		tc("TLSA change matchingtype", tlsa("_443._tcp", 2, 0, 2, sha512hash)),
		tc("TLSA change certificate", tlsa("_443._tcp", 2, 0, 2, reversedSha512)),

		// TXTMulti
		reset(requires(providers.CanUseTXTMulti)),
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
		tc("3x255-byte TXTMulti",
			txtmulti("foo3", []string{strings.Repeat("X", 255), strings.Repeat("Y", 255), strings.Repeat("Z", 255)})),

		//
		// Pseudo rtypes:
		//

		// ALIAS
		reset(requires(providers.CanUseAlias)),
		tc("ALIAS at root", alias("@", "foo.com.")),
		tc("change it", alias("@", "foo2.com.")),
		tc("ALIAS at subdomain", alias("test", "foo.com.")),

		// AZURE_ALIAS
		reset(requires(providers.CanUseAzureAlias)),
		tc("create dependent A records", a("foo.a", "1.2.3.4"), a("quux.a", "2.3.4.5")),
		tc("ALIAS to A record in same zone", a("foo.a", "1.2.3.4"), a("quux.a", "2.3.4.5"), azureAlias("bar.a", "A", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain-no-trailing**/A/foo.a")),
		tc("change it", a("foo.a", "1.2.3.4"), a("quux.a", "2.3.4.5"), azureAlias("bar.a", "A", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain-no-trailing**/A/quux.a")),
		tc("create dependent CNAME records", cname("foo.cname", "google.com"), cname("quux.cname", "google2.com")),
		tc("ALIAS to CNAME record in same zone", cname("foo.cname", "google.com"), cname("quux.cname", "google2.com"), azureAlias("bar", "CNAME", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain-no-trailing**/CNAME/foo.cname")),
		tc("change it", cname("foo.cname", "google.com"), cname("quux.cname", "google2.com"), azureAlias("bar.cname", "CNAME", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain-no-trailing**/CNAME/quux.cname")),

		// R53_ALIAS
		reset(requires(providers.CanUseRoute53Alias)),
		tc("create dependent records", a("foo", "1.2.3.4"), a("quux", "2.3.4.5")),
		tc("ALIAS to A record in same zone", a("foo", "1.2.3.4"), a("quux", "2.3.4.5"), r53alias("bar", "A", "foo.**current-domain**")),
		tc("change it", a("foo", "1.2.3.4"), a("quux", "2.3.4.5"), r53alias("bar", "A", "quux.**current-domain**")),

		//
		// End
		//

		// Close out the previous test.
		reset(),
	}

	return tests
}
