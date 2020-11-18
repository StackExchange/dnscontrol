package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/miekg/dns/dnsutil"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/nameservers"
	"github.com/StackExchange/dnscontrol/v3/providers"
	_ "github.com/StackExchange/dnscontrol/v3/providers/_all"
	"github.com/StackExchange/dnscontrol/v3/providers/config"
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
	t.Run(domain, func(t *testing.T) {
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
func testPermitted(t *testing.T, p string, f TestGroup) error {

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
		return fmt.Errorf("disabled by only")
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

//func makeClearFilter() *TestCase {
//	tc := tc("Empty")
//	tc.ChangeFilter = true
//	return tc
//}

// desc := fmt.Sprintf("%d: %s", i, tst.Desc)
// makeChanges runs one set of DNS record tests. Returns true on success.
func makeChanges(t *testing.T, prv providers.DNSServiceProvider, dc *models.DomainConfig, tst *TestCase, desc string, expectChanges bool, origConfig map[string]string) bool {
	domainName := dc.Name

	return t.Run(desc+":"+tst.Desc, func(t *testing.T) {
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
			//if providers.ProviderHasCapability(*providerToRun, providers.CanUseAzureAlias) {
			if strings.Contains(rc.GetTargetField(), "**subscription-id**") {
				_ = rc.SetTarget(strings.Replace(rc.GetTargetField(), "**subscription-id**", origConfig["SubscriptionID"], 1))
			}
			if strings.Contains(rc.GetTargetField(), "**resource-group**") {
				_ = rc.SetTarget(strings.Replace(rc.GetTargetField(), "**resource-group**", origConfig["ResourceGroup"], 1))
			}
			//}
			dom.Records = append(dom.Records, &rc)
		}
		dom.IgnoredNames = tst.IgnoredNames
		dom.IgnoredTargets = tst.IgnoredTargets
		models.PostProcessRecords(dom.Records)
		dom2, _ := dom.Copy()

		// get and run corrections for first time
		corrections, err := prv.GetDomainCorrections(dom)
		if err != nil {
			t.Fatal(fmt.Errorf("runTests: %w", err))
		}
		if len(corrections) == 0 && expectChanges {
			t.Fatalf("Expected changes, but got none")
		}
		for _, c := range corrections {
			if *verbose {
				t.Log(c.Msg)
			}
			err = c.F()
			if err != nil {
				t.Fatal(err)
			}
		}

		// If we just emptied out the zone, no need for a second pass.
		if len(tst.Records) == 0 {
			return
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

func runTests(t *testing.T, prv providers.DNSServiceProvider, domainName string, knownFailures map[int]bool, origConfig map[string]string) {
	dc := getDomainConfigWithNameservers(t, prv, domainName)
	testGroups := makeTests(t)

	firstGroup := *startIdx
	lastGroup := *endIdx
	if lastGroup == 0 {
		lastGroup = len(testGroups)
	}

	// Start the zone with a clean slate.
	makeChanges(t, prv, dc, tc("Empty"), "Clean Slate", false, nil)

	curGroup := -1
	for gIdx, group := range testGroups {

		// Abide by -start -end flags
		curGroup++
		if curGroup < firstGroup || curGroup > lastGroup {
			continue
		}

		// Abide by filter
		if err := testPermitted(t, *providerToRun, *group); err != nil {
			//t.Logf("%s: ***SKIPPED(%v)***", group.Desc, err)
			makeChanges(t, prv, dc, tc("Empty"), fmt.Sprintf("%02d:%s ***SKIPPED(%v)***", gIdx, group.Desc, err), false, origConfig)
			continue
		}

		// Run the tests.
		for _, tst := range group.tests {
			makeChanges(t, prv, dc, tst, fmt.Sprintf("%02d:%s", gIdx, group.Desc), true, origConfig)
			if t.Failed() {
				break
			}
		}

		// Remove all records so next group starts with a clean slate.
		makeChanges(t, prv, dc, tc("Empty"), "Post cleanup", false, nil)

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

type TestGroup struct {
	Desc     string
	required []providers.Capability
	only     []string
	not      []string
	tests    []*TestCase
}

type TestCase struct {
	Desc           string
	Records        []*rec
	IgnoredNames   []string
	IgnoredTargets []*models.IgnoreTarget
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

func ds(name string, keyTag uint16, algorithm, digestType uint8, digest string) *rec {
	r := makeRec(name, "", "DS")
	r.DsKeyTag = keyTag
	r.DsAlgorithm = algorithm
	r.DsDigestType = digestType
	r.DsDigest = digest
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

func ignoreName(name string) *rec {
	r := &rec{
		Type: "IGNORE_NAME",
	}
	r.SetLabel(name, "**current-domain**")
	return r
}

func ignoreTarget(name string, typ string) *rec {
	r := &rec{
		Type:   "IGNORE_TARGET",
		Target: typ,
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

func testgroup(desc string, items ...interface{}) *TestGroup {
	group := &TestGroup{Desc: desc}
	for _, item := range items {
		switch v := item.(type) {
		case requiresFilter:
			if len(group.tests) != 0 {
				fmt.Printf("ERROR: requires() must be before all tc(): %v\n", desc)
				os.Exit(1)
			}
			group.required = append(group.required, v.caps...)
		case notFilter:
			if len(group.tests) != 0 {
				fmt.Printf("ERROR: not() must be before all tc(): %v\n", desc)
				os.Exit(1)
			}
			group.not = append(group.not, v.names...)
		case onlyFilter:
			if len(group.tests) != 0 {
				fmt.Printf("ERROR: only() must be before all tc(): %v\n", desc)
				os.Exit(1)
			}
			group.only = append(group.only, v.names...)
		case *TestCase:
			group.tests = append(group.tests, v)
		default:
			fmt.Printf("I don't know about type %T (%v)\n", v, v)
		}
	}
	return group
}

func tc(desc string, recs ...*rec) *TestCase {
	var records []*rec
	var ignoredNames []string
	var ignoredTargets []*models.IgnoreTarget
	for _, r := range recs {
		if r.Type == "IGNORE_NAME" {
			ignoredNames = append(ignoredNames, r.GetLabel())
		} else if r.Type == "IGNORE_TARGET" {
			ignoredTargets = append(ignoredTargets, &models.IgnoreTarget{
				Pattern: r.GetLabel(),
				Type:    r.Target,
			})
		} else {
			records = append(records, r)
		}
	}
	return &TestCase{
		Desc:           desc,
		Records:        records,
		IgnoredNames:   ignoredNames,
		IgnoredTargets: ignoredTargets,
	}
}

func clear(items ...interface{}) *TestCase {
	return tc("Empty")
}

type requiresFilter struct {
	caps []providers.Capability
}

func requires(c ...providers.Capability) requiresFilter {
	return requiresFilter{caps: c}
}

type notFilter struct {
	names []string
}

func not(n ...string) notFilter {
	return notFilter{names: n}
}

type onlyFilter struct {
	names []string
}

func only(n ...string) onlyFilter {
	return onlyFilter{names: n}
}

//

func makeTests(t *testing.T) []*TestGroup {

	sha256hash := strings.Repeat("0123456789abcdef", 4)
	sha512hash := strings.Repeat("0123456789abcdef", 8)
	reversedSha512 := strings.Repeat("fedcba9876543210", 8)

	// Each group of tests begins with testgroup("Title").
	// The system will remove any records so that the tests
	// begin with a clean slate (i.e. no records).

	// Filters:

	// Only apply to providers that CanUseAlias.
	//      requires(providers.CanUseAlias),
	// Only apply to ROUTE53 + GANDI_V5:
	//      only("ROUTE53", "GANDI_V5")
	// Only apply to all providers except ROUTE53 + GANDI_V5:
	//     not("ROUTE53", "GANDI_V5"),
	// NOTE: You can't mix not() and only()
	//     reset(not("ROUTE53"), only("GCLOUD")),  // ERROR!
	// NOTE: All requires()/not()/only() must appear before any tc().

	// tc()
	// Each tc() indicates a set of records.  The testgroup tries to
	// migrate from one tc() to the next.  For example the first tc()
	// creates some records. The next tc() might list the same records
	// but adds 1 new record and omits 1.  Therefore migrating to this
	// second tc() results in 1 record being created and 1 deleted; but
	// for some providers it may be converting 1 record to another.
	// Therefore some testgroups are testing the providers ability to
	// transition between different states. Others are just testing
	// whether or not a certain kind of record can be created and
	// deleted.

	// clear() is the same as tc("Empty").  It removes all records.  You
	// can use this to verify a provider can delete all the records in
	// the last tc(), or to provide a clean slate for the next tc().
	// Each testgroup() begins and ends with clear(), so you don't have
	// to list the clear() yourself.

	tests := []*TestGroup{

		//
		// Basic functionality (add/rename/change/delete).
		//

		testgroup("GeneralACD",
			// Test general ability to add/change/delete records of one
			// type. These tests aren't specific to "A" records, but we
			// don't do tests specific to A records because this exercises
			// them very well.
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
		),

		testgroup("WildcardACD",
			not("HEDNS"), // Not supported by dns.he.net due to abuse
			tc("Create wildcard", a("*", "1.2.3.4"), a("www", "1.1.1.1")),
			tc("Delete wildcard", a("www", "1.1.1.1")),
		),

		//
		// Test the basic rtypes.
		//

		testgroup("CNAME",
			tc("Create a CNAME", cname("foo", "google.com.")),
			tc("Change CNAME target", cname("foo", "google2.com.")),
			clear(),
			tc("Record pointing to @", cname("foo", "**current-domain**")),
		),

		testgroup("MX",
			not("ACTIVEDIRECTORY_PS"), // Not implemented.
			tc("MX record", mx("@", 5, "foo.com.")),
			tc("Second MX record, same prio", mx("@", 5, "foo.com."), mx("@", 5, "foo2.com.")),
			tc("3 MX", mx("@", 5, "foo.com."), mx("@", 5, "foo2.com."), mx("@", 15, "foo3.com.")),
			tc("Delete one", mx("@", 5, "foo2.com."), mx("@", 15, "foo3.com.")),
			tc("Change to other name", mx("@", 5, "foo2.com."), mx("mail", 15, "foo3.com.")),
			tc("Change Preference", mx("@", 7, "foo2.com."), mx("mail", 15, "foo3.com.")),
			tc("Record pointing to @", mx("foo", 8, "**current-domain**")),
		),

		testgroup("Null MX",
			not("AZURE_DNS", "GANDI_V5", "INWX", "NAMEDOTCOM", "DIGITALOCEAN", "NETCUP", "DNSIMPLE", "HEDNS", "VULTR"), // These providers don't support RFC 7505
			tc("Null MX", mx("@", 0, ".")),
		),

		testgroup("NS",
			not("DNSIMPLE", "EXOSCALE", "NETCUP"),
			// DNSIMPLE: Does not support NS records nor subdomains.
			// EXOSCALE: FILL IN
			// Netcup: NS records not currently supported.
			tc("NS for subdomain", ns("xyz", "ns2.foo.com.")),
			tc("Dual NS for subdomain", ns("xyz", "ns2.foo.com."), ns("xyz", "ns1.foo.com.")),
			tc("NS Record pointing to @", ns("foo", "**current-domain**")),
		),

		testgroup("IGNORE_NAME function",
			tc("Create some records", txt("foo", "simple"), a("foo", "1.2.3.4")),
			tc("Add a new record - ignoring foo", a("bar", "1.2.3.4"), ignoreName("foo")),
			clear(),
			tc("Create some records", txt("bar.foo", "simple"), a("bar.foo", "1.2.3.4")),
			tc("Add a new record - ignoring *.foo", a("bar", "1.2.3.4"), ignoreName("*.foo")),
		),

		testgroup("IGNORE_TARGET function",
			tc("Create some records", cname("foo", "test.foo.com."), cname("bar", "test.bar.com.")),
			tc("Add a new record - ignoring test.foo.com.", cname("bar", "bar.foo.com."), ignoreTarget("test.foo.com.", "CNAME")),
			clear(),
			tc("Create some records", cname("bar.foo", "a.b.foo.com."), a("test.foo", "1.2.3.4")),
			tc("Add a new record - ignoring **.foo.com. targets", a("bar", "1.2.3.4"), ignoreTarget("**.foo.com.", "CNAME")),
		),

		testgroup("single TXT",
			tc("Create a TXT", txt("foo", "simple")),
			tc("Change a TXT", txt("foo", "changed")),
			clear(),
			tc("Create a TXT with spaces", txt("foo", "with spaces")),
			tc("Create 1 TXT as array", txtmulti("foo", []string{"simple"})), // Same as non-TXTMulti
			clear(),
			tc("Create a 255-byte TXT", txt("foo", strings.Repeat("A", 255))),
			clear(),
			tc("Create TXT with single-quote", txt("foo", "blah`blah")),
		),

		testgroup("ws TXT",
			not("CLOUDFLAREAPI", "HEXONET", "INWX", "NAMEDOTCOM"),
			// These providers strip whitespace at the end of TXT records.
			// TODO(tal): Add a check for this in normalize/validate.go
			tc("Change a TXT with ws at end", txt("foo", "with space at end  ")),
		),

		testgroup("empty TXT",
			not("CLOUDFLAREAPI", "HETZNER", "HEXONET", "INWX", "NETCUP"),
			tc("TXT with empty str", txt("foo1", "")),
			// https://github.com/StackExchange/dnscontrol/issues/598
			// We decided that permitting the TXT target to be an empty
			// string is not a requirement (even though RFC1035 permits it).
			// In the future we might make it "capability" to
			// indicate which vendors support an empty TXT record.
			// However at this time there is no pressing need for this
			// feature.
		),

		//
		// Tests that exercise the API protocol and/or code.
		//

		testgroup("TypeChange",
			// Test whether the provider properly handles a label changing
			// from one rtype to another.
			tc("Create a CNAME", cname("foo", "google.com.")),
			tc("Change to A record", a("foo", "1.2.3.4")),
			tc("Change back to CNAME", cname("foo", "google2.com.")),
		),

		testgroup("Case Sensitivity",
			// The decoys are required so that there is at least one actual change in each tc.
			tc("Create CAPS", mx("BAR", 5, "BAR.com.")),
			tc("Downcase label", mx("bar", 5, "BAR.com."), a("decoy", "1.1.1.1")),
			tc("Downcase target", mx("bar", 5, "bar.com."), a("decoy", "2.2.2.2")),
			tc("Upcase both", mx("BAR", 5, "BAR.COM."), a("decoy", "3.3.3.3")),
		),

		testgroup("IDNA",
			not("SOFTLAYER", "CLOUDFLAREAPI"),
			// SOFTLAYER: fails at direct internationalization, punycode works, of course.
			// CLOUDFLAREAPI: fails. Needs to be debugged.
			tc("Internationalized name", a("ööö", "1.2.3.4")),
			tc("Change IDN", a("ööö", "2.2.2.2")),
			tc("Internationalized CNAME Target", cname("a", "ööö.com.")),
		),
		testgroup("IDNAs in CNAME targets",
			not("LINODE", "CLOUDFLAREAPI"),
			// LINODE: hostname validation does not allow the target domain TLD
			tc("IDN CNAME AND Target", cname("öoö", "ööö.企业.")),
		),

		testgroup("pager101",
			// Tests the paging code of providers.  Many providers page at 100.
			// Notes:
			//  - Gandi: page size is 100, therefore we test with 99, 100, and 101
			//  - NS1: free acct only allows 50 records, therefore we skip
			//  - DIGITALOCEAN: page size is 100 (default: 20)
			//  - CLOUDFLAREAPI: Infinite pagesize but due to slow speed, skipping.
			not("NS1", "CLOUDFLAREAPI"),
			tc("99 records", manyA("rec%04d", "1.2.3.4", 99)...),
			tc("100 records", manyA("rec%04d", "1.2.3.4", 100)...),
			tc("101 records", manyA("rec%04d", "1.2.3.4", 101)...),
		),

		testgroup("pager601",
			// AWS:  https://github.com/StackExchange/dnscontrol/issues/493
			//only("AZURE_DNS", "HEXONET", "ROUTE53"),
			only("HEXONET", "ROUTE53", "GCLOUD"), // AZURE_DNS is failing.
			tc("601 records", manyA("rec%04d", "1.2.3.4", 600)...),
			tc("Update 601 records", manyA("rec%04d", "1.2.3.5", 600)...),
		),

		testgroup("pager1201",
			// AWS:  https://github.com/StackExchange/dnscontrol/issues/493
			// Azure: https://github.com/StackExchange/dnscontrol/issues/770
			//only("AZURE_DNS", "HEXONET", "ROUTE53"),
			only("HEXONET", "ROUTE53"), // AZURE_DNS is failing.
			tc("1200 records", manyA("rec%04d", "1.2.3.4", 1200)...),
			tc("Update 1200 records", manyA("rec%04d", "1.2.3.5", 1200)...),
		),

		//
		// CanUse* types:
		//

		testgroup("CAA",
			requires(providers.CanUseCAA),
			tc("CAA record", caa("@", "issue", 0, "letsencrypt.org")),
			tc("CAA change tag", caa("@", "issuewild", 0, "letsencrypt.org")),
			tc("CAA change target", caa("@", "issuewild", 0, "example.com")),
			tc("CAA change flag", caa("@", "issuewild", 128, "example.com")),
			tc("CAA many records",
				caa("@", "issue", 0, "letsencrypt.org"),
				caa("@", "issuewild", 0, "comodoca.com"),
				caa("@", "iodef", 128, "mailto:test@example.com")),
			tc("CAA delete", caa("@", "issue", 0, "letsencrypt.org")),
		),
		testgroup("CAA with ;",
			requires(providers.CanUseCAA), not("DIGITALOCEAN"),
			// Test support of ";" as a value
			tc("CAA many records", caa("@", "issuewild", 0, ";")),
		),

		testgroup("NAPTR",
			requires(providers.CanUseNAPTR),
			tc("NAPTR record", naptr("test", 100, 10, "U", "E2U+sip", "!^.*$!sip:customer-service@example.com!", "example.foo.com.")),
			tc("NAPTR second record", naptr("test", 102, 10, "U", "E2U+email", "!^.*$!mailto:information@example.com!", "example.foo.com.")),
			tc("NAPTR delete record", naptr("test", 100, 10, "U", "E2U+email", "!^.*$!mailto:information@example.com!", "example.foo.com.")),
			tc("NAPTR change target", naptr("test", 100, 10, "U", "E2U+email", "!^.*$!mailto:information@example.com!", "example2.foo.com.")),
			tc("NAPTR change order", naptr("test", 103, 10, "U", "E2U+email", "!^.*$!mailto:information@example.com!", "example2.foo.com.")),
			tc("NAPTR change preference", naptr("test", 103, 20, "U", "E2U+email", "!^.*$!mailto:information@example.com!", "example2.foo.com.")),
			tc("NAPTR change flags", naptr("test", 103, 20, "A", "E2U+email", "!^.*$!mailto:information@example.com!", "example2.foo.com.")),
			tc("NAPTR change service", naptr("test", 103, 20, "A", "E2U+sip", "!^.*$!mailto:information@example.com!", "example2.foo.com.")),
			tc("NAPTR change regexp", naptr("test", 103, 20, "A", "E2U+sip", "!^.*$!sip:customer-service@example.com!", "example2.foo.com.")),
		),

		testgroup("PTR", requires(providers.CanUsePTR), not("ACTIVEDIRECTORY_PS"),
			tc("Create PTR record", ptr("4", "foo.com.")),
			tc("Modify PTR record", ptr("4", "bar.com.")),
		),

		testgroup("SRV", requires(providers.CanUseSRV), not("ACTIVEDIRECTORY_PS", "CLOUDNS"),
			tc("SRV record", srv("_sip._tcp", 5, 6, 7, "foo.com.")),
			tc("Second SRV record, same prio", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 5, 60, 70, "foo2.com.")),
			tc("3 SRV", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 5, 60, 70, "foo2.com."), srv("_sip._tcp", 15, 65, 75, "foo3.com.")),
			tc("Delete one", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo3.com.")),
			tc("Change Target", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
			tc("Change Priority", srv("_sip._tcp", 52, 6, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
			tc("Change Weight", srv("_sip._tcp", 52, 62, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
			tc("Change Port", srv("_sip._tcp", 52, 62, 72, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
		),
		testgroup("SRV w/ null target", requires(providers.CanUseSRV), not("EXOSCALE", "HEXONET", "INWX", "NAMEDOTCOM"),
			tc("Null Target", srv("_sip._tcp", 52, 62, 72, "foo.com."), srv("_sip._tcp", 15, 65, 75, ".")),
		),

		testgroup("SSHFP",
			requires(providers.CanUseSSHFP),
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
		),

		testgroup("TLSA",
			requires(providers.CanUseTLSA),
			tc("TLSA record", tlsa("_443._tcp", 3, 1, 1, sha256hash)),
			tc("TLSA change usage", tlsa("_443._tcp", 2, 1, 1, sha256hash)),
			tc("TLSA change selector", tlsa("_443._tcp", 2, 0, 1, sha256hash)),
			tc("TLSA change matchingtype", tlsa("_443._tcp", 2, 0, 2, sha512hash)),
			tc("TLSA change certificate", tlsa("_443._tcp", 2, 0, 2, reversedSha512)),
		),

		testgroup("TXTMulti",
			requires(providers.CanUseTXTMulti),
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
			tc("Change TXTMulti",
				txtmulti("foo1", []string{"dimple"}),
				txtmulti("foo2", []string{"fun", "two"}),
				txtmulti("foo3", []string{"eh", "bzz", "cee"}),
			),
			tc("Long TXTMulti",
				// This isn't the longest a TXT record can be, but it is the
				// longest DIGITALOCEAN permits.
				txtmulti("foo4", []string{strings.Repeat("X", 255), strings.Repeat("Y", 252)}),
			),
		),

		testgroup("TXTMulti tests that break DO",
			// DO's implementation of TXTMulti is broken, thus we separate out
			// tests that fail for it. Users are warned about these limits
			// in docs/_providers/digitalocean.md
			requires(providers.CanUseTXTMulti),
			not("DIGITALOCEAN"),
			// Digital Ocean's TXT record implementation checks size limits wrong.
			// RFC 1035 Section 3.3.14 states that each substring can be 255
			// octets, and there is no limit on the number of such
			// substrings, aside from the usual packet length limits.  DO's
			// implementation restricts the total length to be 512 octets,
			// including any backlashes used for escapes, quotes, and other
			// metachars.  In other words, they're doing the checking on the
			// API protocol encoded data instead of on on the resulting TXT
			// record.
			// Proper TXT implementations can handle TXT records like this:
			tc("3x255-byte TXTMulti",
				txtmulti("foo3", []string{strings.Repeat("X", 255), strings.Repeat("Y", 255), strings.Repeat("Z", 255)})),
			clear(),
			// Digital Ocean's TXT record implementation handles quotes wrong.
			// It craps out if your TXT record includes double-quotes.
			// Someone doesn't understand how zonefile-style quoting is
			// supposed to work.
			// Proper TXT implementations can handle TXT records like these:
			tc("Create TXTMulti with quotes",
				txtmulti("foo1", []string{"simple"}),
				txtmulti("foo2", []string{"o\"ne", "tw\"o"}),
				txtmulti("foo3", []string{"eh", "bee", "cee"}),
			),
			tc("Change TXTMulti",
				txtmulti("foo1", []string{"dimple"}),
				txtmulti("foo2", []string{"fun", "t\"wo"}),
				txtmulti("foo3", []string{"eh", "bzz", "cee"}),
			),
		),

		testgroup("DS",
			requires(providers.CanUseDS),
			tc("create DS", ds("@", 1, 13, 1, "ADIGEST")),
			tc("modify field 1", ds("@", 65535, 13, 1, "ADIGEST")),
			tc("modify field 3", ds("@", 65535, 13, 2, "ADIGEST")),
			tc("modify field 2+3", ds("@", 65535, 1, 4, "ADIGEST")),
			tc("modify field 2", ds("@", 65535, 3, 4, "ADIGEST")),
			tc("modify field 2", ds("@", 65535, 254, 4, "ADIGEST")),
			tc("delete 1, create 1", ds("foo", 2, 13, 4, "ADIGEST")),
			tc("add 2 more DS", ds("foo", 2, 13, 4, "ADIGEST"), ds("@", 65535, 5, 4, "ADIGEST"), ds("@", 65535, 253, 4, "ADIGEST")),
		),

		testgroup("DS (children only)",
			requires(providers.CanUseDSForChildren),
			// Use a valid digest value here, because GCLOUD (which implements this capability) verifies
			// the value passed in is a valid digest. RFC 4034, s5.1.4 specifies SHA1 as the only digest
			// algo at present, i.e. only hexadecimal values currently usable.
			tc("create DS", ds("child", 1, 13, 1, "0123456789ABCDEF")),
			tc("modify field 1", ds("child", 65535, 13, 1, "0123456789ABCDEF")),
			tc("modify field 3", ds("child", 65535, 13, 2, "0123456789ABCDEF")),
			tc("modify field 2+3", ds("child", 65535, 1, 4, "0123456789ABCDEF")),
			tc("modify field 2", ds("child", 65535, 3, 4, "0123456789ABCDEF")),
			tc("modify field 2", ds("child", 65535, 254, 4, "0123456789ABCDEF")),
			tc("delete 1, create 1", ds("another-child", 2, 13, 4, "0123456789ABCDEF")),
			tc("add 2 more DS",
				ds("another-child", 2, 13, 4, "0123456789ABCDEF"),
				ds("another-child", 65535, 5, 4, "0123456789ABCDEF"),
				ds("another-child", 65535, 253, 4, "0123456789ABCDEF"),
			),
		),

		//
		// Pseudo rtypes:
		//

		testgroup("ALIAS",
			requires(providers.CanUseAlias),
			tc("ALIAS at root", alias("@", "foo.com.")),
			tc("change it", alias("@", "foo2.com.")),
			tc("ALIAS at subdomain", alias("test", "foo.com.")),
			tc("change it", alias("test", "foo2.com.")),
		),

		testgroup("AZURE_ALIAS",
			requires(providers.CanUseAzureAlias),
			tc("create dependent A records",
				a("foo.a", "1.2.3.4"),
				a("quux.a", "2.3.4.5"),
			),
			tc("ALIAS to A record in same zone",
				a("foo.a", "1.2.3.4"),
				a("quux.a", "2.3.4.5"),
				azureAlias("bar.a", "A", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain-no-trailing**/A/foo.a"),
			),
			tc("change it",
				a("foo.a", "1.2.3.4"),
				a("quux.a", "2.3.4.5"),
				azureAlias("bar.a", "A", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain-no-trailing**/A/quux.a"),
			),
			tc("create dependent CNAME records",
				cname("foo.cname", "google.com"),
				cname("quux.cname", "google2.com"),
			),
			tc("ALIAS to CNAME record in same zone",
				cname("foo.cname", "google.com"),
				cname("quux.cname", "google2.com"),
				azureAlias("bar", "CNAME", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain-no-trailing**/CNAME/foo.cname"),
			),
			tc("change it",
				cname("foo.cname", "google.com"),
				cname("quux.cname", "google2.com"),
				azureAlias("bar.cname", "CNAME", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain-no-trailing**/CNAME/quux.cname"),
			),
		),

		testgroup("R53_ALIAS2",
			requires(providers.CanUseRoute53Alias),
			tc("create dependent records",
				a("kyle", "1.2.3.4"),
				a("cartman", "2.3.4.5"),
			),
			tc("ALIAS to A record in same zone",
				a("kyle", "1.2.3.4"),
				a("cartman", "2.3.4.5"),
				r53alias("kenny", "A", "kyle.**current-domain**"),
			),
			tc("modify an r53 alias",
				a("kyle", "1.2.3.4"),
				a("cartman", "2.3.4.5"),
				r53alias("kenny", "A", "cartman.**current-domain**"),
			),
		),

		testgroup("R53_ALIAS_ORDER",
			requires(providers.CanUseRoute53Alias),
			tc("create target cnames",
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
				cname("dev-system19", "ec2-54-91-99-999.compute-1.amazonaws.com."),
			),
			tc("add an alias to 18",
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
				cname("dev-system19", "ec2-54-91-99-999.compute-1.amazonaws.com."),
				r53alias("dev-system", "CNAME", "dev-system18.**current-domain**"),
			),
			tc("modify alias to 19",
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
				cname("dev-system19", "ec2-54-91-99-999.compute-1.amazonaws.com."),
				r53alias("dev-system", "CNAME", "dev-system19.**current-domain**"),
			),
			tc("remove alias",
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
				cname("dev-system19", "ec2-54-91-99-999.compute-1.amazonaws.com."),
			),
			tc("add an alias back",
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
				cname("dev-system19", "ec2-54-91-99-999.compute-1.amazonaws.com."),
				r53alias("dev-system", "CNAME", "dev-system19.**current-domain**"),
			),
			tc("remove cnames",
				r53alias("dev-system", "CNAME", "dev-system19.**current-domain**"),
			),
			clear(),
			tc("create cname+alias in one step",
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
				r53alias("dev-system", "CNAME", "dev-system18.**current-domain**"),
			),
			clear(),
			tc("create alias+cname in one step",
				r53alias("dev-system", "CNAME", "dev-system18.**current-domain**"),
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
			),
		),
	}

	return tests
}
