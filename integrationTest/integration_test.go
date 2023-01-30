package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/credsfile"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v3/pkg/nameservers"
	"github.com/StackExchange/dnscontrol/v3/pkg/normalize"
	"github.com/StackExchange/dnscontrol/v3/providers"
	_ "github.com/StackExchange/dnscontrol/v3/providers/_all"
	"github.com/StackExchange/dnscontrol/v3/providers/cloudflare"
	"github.com/miekg/dns/dnsutil"
)

var providerToRun = flag.String("provider", "", "Provider to run")
var startIdx = flag.Int("start", 0, "Test number to begin with")
var endIdx = flag.Int("end", 0, "Test index to stop after")
var verbose = flag.Bool("verbose", false, "Print corrections as you run them")
var printElapsed = flag.Bool("elapsed", false, "Print elapsed time for each testgroup")
var enableCFWorkers = flag.Bool("cfworkers", true, "Set false to disable CF worker tests")

func init() {
	testing.Init()

	flag.BoolVar(&diff2.EnableDiff2, "diff2", false, "enable diff2")
	flag.Parse()
}

func getProvider(t *testing.T) (providers.DNSServiceProvider, string, map[int]bool, map[string]string) {
	if *providerToRun == "" {
		t.Log("No provider specified with -provider")
		return nil, "", nil, nil
	}
	jsons, err := credsfile.LoadProviderConfigs("providers.json")
	if err != nil {
		t.Fatalf("Error loading provider configs: %s", err)
	}
	fails := map[int]bool{}
	for name, cfg := range jsons {
		if *providerToRun != name {
			continue
		}

		var metadata json.RawMessage
		// CLOUDFLAREAPI tests related to CF_REDIRECT/CF_TEMP_REDIRECT
		// requires metadata to enable this feature.
		// In hindsight, I have no idea why this metadata flag is required to
		// use this feature. Maybe because we didn't have the capabilities
		// feature at the time?
		if name == "CLOUDFLAREAPI" {
			if *enableCFWorkers {
				metadata = []byte(`{ "manage_redirects": true, "manage_workers": true }`)
			} else {
				metadata = []byte(`{ "manage_redirects": true }`)
			}
		}

		provider, err := providers.CreateDNSProvider(name, cfg, metadata)
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

		if name == "CLOUDFLAREAPI" && *enableCFWorkers {
			// Cloudflare only. Will do nothing if provider != *cloudflareProvider.
			if err := cloudflare.PrepareCloudflareTestWorkers(provider); err != nil {
				t.Fatal(err)
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
	if domain == "" {
		t.Fatal("NO DOMAIN SET!  Exiting!")
	}

	t.Run(domain, func(t *testing.T) {
		runTests(t, provider, domain, fails, cfg)
	})

}

func getDomainConfigWithNameservers(t *testing.T, prv providers.DNSServiceProvider, domainName string) *models.DomainConfig {
	dc := &models.DomainConfig{
		Name: domainName,
	}
	normalize.UpdateNameSplitHorizon(dc)

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

	// If there are any trueflags, make sure they are all true.
	for _, c := range f.trueflags {
		if !c {
			return fmt.Errorf("excluded by alltrue(%v)", f.trueflags)
		}
	}

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

		if err := providers.AuditRecords(*providerToRun, dom.Records); err != nil {
			t.Skipf("***SKIPPED(PROVIDER DOES NOT SUPPORT '%s' ::%q)", err, desc)
			return
		}

		// get and run corrections for first time
		corrections, err := prv.GetDomainCorrections(dom)
		if err != nil {
			t.Fatal(fmt.Errorf("runTests: %w", err))
		}
		if (len(corrections) == 0 && expectChanges) && (tst.Desc != "Empty") {
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
				t.Logf("UNEXPECTED #%d: %s", i, c.Msg)
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
		start := time.Now()

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

		elapsed := time.Since(start)
		if *printElapsed {
			fmt.Printf("ELAPSED %02d %7.2f %q\n", gIdx, elapsed.Seconds(), group.Desc)
		}

	}

}

func TestDualProviders(t *testing.T) {
	p, domain, _, _ := getProvider(t)
	if p == nil {
		return
	}
	if domain == "" {
		t.Fatal("NO DOMAIN SET!  Exiting!")
	}
	dc := getDomainConfigWithNameservers(t, p, domain)
	if !providers.ProviderHasCapability(*providerToRun, providers.DocDualHost) {
		t.Skip("Skipping.  DocDualHost == Cannot")
		return
	}
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
	nslist, _ := models.ToNameservers([]string{"ns1.example.com", "ns2.example.com"})
	dc.Nameservers = append(dc.Nameservers, nslist...)
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
	Desc      string
	required  []providers.Capability
	only      []string
	not       []string
	trueflags []bool
	tests     []*TestCase
}

type TestCase struct {
	Desc           string
	Records        []*models.RecordConfig
	IgnoredNames   []*models.IgnoreName
	IgnoredTargets []*models.IgnoreTarget
}

func SetLabel(r *models.RecordConfig, label, domain string) {
	r.Name = label
	r.NameFQDN = dnsutil.AddOrigin(label, "**current-domain**")
}

func a(name, target string) *models.RecordConfig {
	return makeRec(name, target, "A")
}

func cname(name, target string) *models.RecordConfig {
	return makeRec(name, target, "CNAME")
}

func alias(name, target string) *models.RecordConfig {
	return makeRec(name, target, "ALIAS")
}

func r53alias(name, aliasType, target string) *models.RecordConfig {
	r := makeRec(name, target, "R53_ALIAS")
	r.R53Alias = map[string]string{
		"type": aliasType,
	}
	return r
}

func azureAlias(name, aliasType, target string) *models.RecordConfig {
	r := makeRec(name, target, "AZURE_ALIAS")
	r.AzureAlias = map[string]string{
		"type": aliasType,
	}
	return r
}

func cfRedir(pattern, target string) *models.RecordConfig {
	t := fmt.Sprintf("%s,%s", pattern, target)
	r := makeRec("@", t, "CF_REDIRECT")
	return r
}

func cfRedirTemp(pattern, target string) *models.RecordConfig {
	t := fmt.Sprintf("%s,%s", pattern, target)
	r := makeRec("@", t, "CF_TEMP_REDIRECT")
	return r
}

func cfProxyA(name, target, status string) *models.RecordConfig {
	r := a(name, target)
	r.Metadata = make(map[string]string)
	r.Metadata["cloudflare_proxy"] = status
	return r
}

func cfProxyCNAME(name, target, status string) *models.RecordConfig {
	r := cname(name, target)
	r.Metadata = make(map[string]string)
	r.Metadata["cloudflare_proxy"] = status
	return r
}

func cfWorkerRoute(pattern, target string) *models.RecordConfig {
	t := fmt.Sprintf("%s,%s", pattern, target)
	r := makeRec("@", t, "CF_WORKER_ROUTE")
	return r
}

func ns(name, target string) *models.RecordConfig {
	return makeRec(name, target, "NS")
}

func mx(name string, prio uint16, target string) *models.RecordConfig {
	r := makeRec(name, target, "MX")
	r.MxPreference = prio
	return r
}

func ptr(name, target string) *models.RecordConfig {
	return makeRec(name, target, "PTR")
}

func naptr(name string, order uint16, preference uint16, flags string, service string, regexp string, target string) *models.RecordConfig {
	r := makeRec(name, target, "NAPTR")
	r.SetTargetNAPTR(order, preference, flags, service, regexp, target)
	return r
}

func ds(name string, keyTag uint16, algorithm, digestType uint8, digest string) *models.RecordConfig {
	r := makeRec(name, "", "DS")
	r.SetTargetDS(keyTag, algorithm, digestType, digest)
	return r
}

func soa(name string, ns, mbox string, serial, refresh, retry, expire, minttl uint32) *models.RecordConfig {
	r := makeRec(name, "", "SOA")
	r.SetTargetSOA(ns, mbox, serial, refresh, retry, expire, minttl)
	return r
}

func srv(name string, priority, weight, port uint16, target string) *models.RecordConfig {
	r := makeRec(name, target, "SRV")
	r.SetTargetSRV(priority, weight, port, target)
	return r
}

func sshfp(name string, algorithm uint8, fingerprint uint8, target string) *models.RecordConfig {
	r := makeRec(name, target, "SSHFP")
	r.SetTargetSSHFP(algorithm, fingerprint, target)
	return r
}

func txt(name, target string) *models.RecordConfig {
	r := makeRec(name, "", "TXT")
	r.SetTargetTXT(target)
	return r
}

func caa(name string, tag string, flag uint8, target string) *models.RecordConfig {
	r := makeRec(name, target, "CAA")
	r.SetTargetCAA(flag, tag, target)
	return r
}

func tlsa(name string, usage, selector, matchingtype uint8, target string) *models.RecordConfig {
	r := makeRec(name, target, "TLSA")
	r.SetTargetTLSA(usage, selector, matchingtype, target)
	return r
}

func urlfwd(name, target string) *models.RecordConfig {
	return makeRec(name, target, "URLFWD")
}

func ignoreName(name string) *models.RecordConfig {
	r := &models.RecordConfig{
		Type: "IGNORE_NAME",
	}
	SetLabel(r, name, "**current-domain**")
	return r
}

func ignoreTarget(name string, typ string) *models.RecordConfig {
	r := &models.RecordConfig{
		Type: "IGNORE_TARGET",
	}
	r.SetTarget(typ)
	SetLabel(r, name, "**current-domain**")
	return r
}

func makeRec(name, target, typ string) *models.RecordConfig {
	r := &models.RecordConfig{
		Type: typ,
		TTL:  300,
	}
	SetLabel(r, name, "**current-domain**")
	r.SetTarget(target)
	return r
}

// func (r *models.RecordConfig) ttl(t uint32) *models.RecordConfig {
func ttl(r *models.RecordConfig, t uint32) *models.RecordConfig {
	r.TTL = t
	return r
}

func manyA(namePattern, target string, n int) []*models.RecordConfig {
	recs := []*models.RecordConfig{}
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
		case alltrueFilter:
			if len(group.tests) != 0 {
				fmt.Printf("ERROR: alltrue() must be before all tc(): %v\n", desc)
				os.Exit(1)
			}
			group.trueflags = append(group.trueflags, v.flags...)
		case *TestCase:
			group.tests = append(group.tests, v)
		default:
			fmt.Printf("I don't know about type %T (%v)\n", v, v)
		}
	}
	return group
}

func tc(desc string, recs ...*models.RecordConfig) *TestCase {
	var records []*models.RecordConfig
	var ignoredNames []*models.IgnoreName
	var ignoredTargets []*models.IgnoreTarget
	for _, r := range recs {
		if r.Type == "IGNORE_NAME" {
			ignoredNames = append(ignoredNames, &models.IgnoreName{Pattern: r.GetLabel(), Types: r.GetTargetField()})
		} else if r.Type == "IGNORE_TARGET" {
			rec := &models.IgnoreTarget{
				Pattern: r.GetLabel(),
				Type:    r.GetTargetField(),
			}
			ignoredTargets = append(ignoredTargets, rec)
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

type alltrueFilter struct {
	flags []bool
}

func alltrue(f ...bool) alltrueFilter {
	return alltrueFilter{flags: f}
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
	// Only run this test if all these bool flags are true:
	//     alltrue(*enableCFWorkers, *anotherFlag, myBoolValue)
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
		// These tests verify the basic operations of the API: Create, Change, Delete.
		// These are tested on "@" and "www".
		// When these tests pass, you've implemented the basics correctly.

		testgroup("Protocol-Plain",
			tc("Create an A record", a("@", "1.1.1.1")),
			tc("Change it", a("@", "1.2.3.4")),
			tc("Add another", a("@", "1.2.3.4"), a("www", "1.2.3.4")),
			tc("Add another(same name)", a("@", "1.2.3.4"), a("www", "1.2.3.4"), a("www", "5.6.7.8")),
		),

		testgroup("Protocol-TTL",
			not("NETCUP"), // NETCUP does not support TTLs.
			tc("Start", a("@", "1.2.3.4"), a("www", "1.2.3.4"), a("www", "5.6.7.8")),
			tc("Change a ttl", ttl(a("@", "1.2.3.4"), 1000), a("www", "1.2.3.4"), a("www", "5.6.7.8")),
			tc("Change single target from set", ttl(a("@", "1.2.3.4"), 1000), a("www", "2.2.2.2"), a("www", "5.6.7.8")),
			tc("Change all ttls", ttl(a("@", "1.2.3.4"), 500), ttl(a("www", "2.2.2.2"), 400), ttl(a("www", "5.6.7.8"), 400)),
			tc("Delete one", ttl(a("@", "1.2.3.4"), 500), ttl(a("www", "5.6.7.8"), 400)),
		),

		testgroup("add to existing label",
			tc("Setup", ttl(a("www", "5.6.7.8"), 400)),
			tc("Add at same label", ttl(a("www", "5.6.7.8"), 400), ttl(a("www", "1.2.3.4"), 400)),
		),

		// This is a strange one.  It adds a new record to an existing
		// label but the pre-existing label has its TTL change.
		testgroup("add to label and change orig ttl",
			tc("Setup", ttl(a("www", "5.6.7.8"), 400)),
			tc("Add at same label, new ttl", ttl(a("www", "5.6.7.8"), 700), ttl(a("www", "1.2.3.4"), 700)),
		),

		testgroup("Protocol-Wildcard",
			// Test the basic Add/Change/Delete with the domain wildcard.
			not("HEDNS"), // Not supported by dns.he.net due to abuse
			tc("Create wildcard", a("*", "1.2.3.4"), a("www", "1.1.1.1")),
			tc("Delete wildcard", a("www", "1.1.1.1")),
		),

		testgroup("TypeChange",
			// Test whether the provider properly handles a label changing
			// from one rtype to another.
			tc("Create a CNAME", cname("foo", "google.com.")),
			tc("Change to A record", a("foo", "1.2.3.4")),
			tc("Change back to CNAME", cname("foo", "google2.com.")),
		),

		//
		// Test each basic DNS type
		//
		// This tests all the common DNS types in parallel for speed.
		// First: 1 of each type is created.
		// Second: the first parameter is modified.
		// Third: the second parameter is modified. (if there is none, no changes)

		// NOTE: Previously we did a seperate test for each type. It was
		// very slow on certain providers. This is faster but is a little
		// more difficult to read.

		testgroup("CommonDNS",
			tc("Create 1 of each",
				//a("testa", "1.1.1.1"),  // Duplicates work done by Protocol-Plain
				cname("testcname", "example.com."),
				mx("testmx", 5, "foo.com."),
				txt("testtxt", "simple"),
			),
			tc("Change param1",
				//a("testa", "2.2.2.2"),  // Duplicates work done by Protocol-Plain
				cname("testcname", "example2.com."),
				mx("testmx", 6, "foo.com."),
				txt("testtxt", "changed"),
			),
			tc("Change param2", // if there is one)
				//a("testa", "2.2.2.2"),  // Duplicates work done by Protocol-Plain
				cname("testcname", "example2.com."),
				mx("testmx", 6, "bar.com."),
				txt("testtxt", "changed"),
			),
		),

		//
		// Test edge cases from various types.
		//

		testgroup("CNAME",
			tc("Record pointing to @", cname("foo", "**current-domain**")),
		),

		testgroup("MX",
			tc("Record pointing to @", mx("foo", 8, "**current-domain**")),
			tc("Null MX", mx("@", 0, ".")), // RFC 7505
		),

		testgroup("NS",
			not(
				"DNSIMPLE", // Does not support NS records nor subdomains.
				"EXOSCALE", // Not supported.
				"NETCUP",   // NS records not currently supported.
			),
			tc("NS for subdomain", ns("xyz", "ns2.foo.com.")),
			tc("Dual NS for subdomain", ns("xyz", "ns2.foo.com."), ns("xyz", "ns1.foo.com.")),
			tc("NS Record pointing to @", a("@", "1.2.3.4"), ns("foo", "**current-domain**")),
		),

		// In this next section we test all the edge cases related to TXT
		// records. Compliance with the RFCs varies greatly with each provider.
		// Rather than creating a "Capability" for each possible different
		// failing or malcompliance (there would be many!), each provider
		// supplies a function AuditRecords() which returns an error if
		// the provider can not support a record.
		// The integration tests use this feedback to skip tests that we know would fail.
		// (Elsewhere the result of AuditRecords() is used in the
		// "dnscontrol check" phase.)

		testgroup("complex TXT",
			// Do not use only()/not()/requires() in this section.
			// If your provider needs to skip one of these tests, update
			// "provider/*/recordaudit.AuditRecords()" to reject that kind
			// of record. When the provider fixes the bug or changes behavior,
			// update the AuditRecords().

			tc("TXT with 0-octel string", txt("foo1", "")),
			// https://github.com/StackExchange/dnscontrol/issues/598
			// RFC1035 permits this, but rarely do provider support it.
			//clear(),

			tc("a 255-byte TXT", txt("foo255", strings.Repeat("C", 255))),
			//clear(),
			tc("a 256-byte TXT", txt("foo256", strings.Repeat("D", 256))),
			//clear(),

			tc("a 512-byte TXT", txt("foo512", strings.Repeat("C", 512))),
			//clear(),
			tc("a 513-byte TXT", txt("foo513", strings.Repeat("D", 513))),
			//clear(),

			tc("TXT with 1 single-quote", txt("foosq", "quo'te")),
			//clear(),
			tc("TXT with 1 backtick", txt("foobt", "blah`blah")),
			//clear(),
			tc("TXT with 1 double-quotes", txt("foodq", `quo"te`)),
			//clear(),
			tc("TXT with 2 double-quotes", txt("foodqs", `q"uo"te`)),
			//clear(),

			tc("a TXT with interior ws", txt("foosp", "with spaces")),
			//clear(),
			tc("TXT with ws at end", txt("foows1", "with space at end ")),
			//clear(),

			//tc("Create a TXT/SPF", txt("foo", "v=spf1 ip4:99.99.99.99 -all")),
			// This was added because Vultr syntax-checks TXT records with SPF contents.
			//clear(),

			// TODO(tlim): Re-add this when we fix the RFC1035 escaped-quotes issue.
			//tc("Create TXT with frequently escaped characters", txt("fooex", `!^.*$@#%^&()([][{}{<></:;-_=+\`)),
		),

		//
		// API Edge Cases
		//

		testgroup("Case Sensitivity",
			// The decoys are required so that there is at least one actual change in each tc.
			tc("Create CAPS", mx("BAR", 5, "BAR.com.")),
			tc("Downcase label", mx("bar", 5, "BAR.com."), a("decoy", "1.1.1.1")),
			tc("Downcase target", mx("bar", 5, "bar.com."), a("decoy", "2.2.2.2")),
			tc("Upcase both", mx("BAR", 5, "BAR.COM."), a("decoy", "3.3.3.3")),
		),

		testgroup("IDNA",
			not("SOFTLAYER"),
			// SOFTLAYER: fails at direct internationalization, punycode works, of course.
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
			//  - DIGITALOCEAN: page size is 100 (default: 20)
			not(
				"AZURE_DNS",     // Removed because it is too slow
				"CLOUDFLAREAPI", // Infinite pagesize but due to slow speed, skipping.
				"CSCGLOBAL",     // Doesn't page. Works fine.  Due to the slow API we skip.
				"GANDI_V5",      // Their API is so damn slow. We'll add it back as needed.
				"MSDNS",         //  No paging done. No need to test.
				"NAMEDOTCOM",    // Their API is so damn slow. We'll add it back as needed.
				"NS1",           // Free acct only allows 50 records, therefore we skip
			),
			tc("99 records", manyA("rec%04d", "1.2.3.4", 99)...),
			tc("100 records", manyA("rec%04d", "1.2.3.4", 100)...),
			tc("101 records", manyA("rec%04d", "1.2.3.4", 101)...),
		),

		testgroup("pager601",
			only(
				//"AZURE_DNS", // Removed because it is too slow
				//"CLOUDFLAREAPI", // Infinite pagesize but due to slow speed, skipping.
				//"CSCGLOBAL", // Doesn't page. Works fine.  Due to the slow API we skip.
				//"GANDI_V5",   // Their API is so damn slow. We'll add it back as needed.
				"GCLOUD",
				"HEXONET",
				//"MSDNS",     //  No paging done. No need to test.
				"ROUTE53",
			),
			tc("601 records", manyA("rec%04d", "1.2.3.4", 600)...),
			tc("Update 601 records", manyA("rec%04d", "1.2.3.5", 600)...),
		),

		testgroup("pager1201",
			only(
				//"AKAMAIEDGEDNS", // No paging done. No need to test.
				//"AZURE_DNS", // Currently failing. See https://github.com/StackExchange/dnscontrol/issues/770
				//"CLOUDFLAREAPI", // Fails with >1000 corrections. See https://github.com/StackExchange/dnscontrol/issues/1440
				//"CSCGLOBAL",     // Doesn't page. Works fine.  Due to the slow API we skip.
				//"GANDI_V5",   // Their API is so damn slow. We'll add it back as needed.
				"HEXONET",
				"HOSTINGDE",
				//"MSDNS",         // No paging done. No need to test.
				"ROUTE53",
			),
			tc("1200 records", manyA("rec%04d", "1.2.3.4", 1200)...),
			tc("Update 1200 records", manyA("rec%04d", "1.2.3.5", 1200)...),
		),

		testgroup("NS1_URLFWD tests",
			only("NS1"),
			tc("Add a urlfwd", urlfwd("urlfwd1", "/ http://example.com 302 2 0")),
			tc("Update a urlfwd", urlfwd("urlfwd1", "/ http://example.org 301 2 0")),
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
			tc("CAA many records", caa("@", "issuewild", 128, ";")),
			// Test support of spaces in the 3rd field. Some providers don't
			// support this.  See providers/exoscale/auditrecords.go as an example.
			tc("CAA whitespace", caa("@", "issue", 0, "letsencrypt.org; validationmethods=dns-01; accounturi=https://acme-v02.api.letsencrypt.org/acme/acct/1234")),
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

		// ClouDNS provider can work with PTR records, but you need to create special type of zone
		testgroup("PTR", requires(providers.CanUsePTR), not("CLOUDNS"),
			tc("Create PTR record", ptr("4", "foo.com.")),
			tc("Modify PTR record", ptr("4", "bar.com.")),
		),

		// SOA
		testgroup("SOA", requires(providers.CanUseSOA),
			clear(), // Extra clear required or only the first run passes.
			tc("Create SOA record", soa("@", "kim.ns.cloudflare.com.", "dns.cloudflare.com.", 2037190000, 10000, 2400, 604800, 3600)),
			tc("Modify SOA ns    ", soa("@", "mmm.ns.cloudflare.com.", "dns.cloudflare.com.", 2037190000, 10000, 2400, 604800, 3600)),
			tc("Modify SOA mbox  ", soa("@", "mmm.ns.cloudflare.com.", "eee.cloudflare.com.", 2037190000, 10000, 2400, 604800, 3600)),
			tc("Modify SOA refres", soa("@", "mmm.ns.cloudflare.com.", "eee.cloudflare.com.", 2037190000, 10001, 2400, 604800, 3600)),
			tc("Modify SOA retry ", soa("@", "mmm.ns.cloudflare.com.", "eee.cloudflare.com.", 2037190000, 10001, 2401, 604800, 3600)),
			tc("Modify SOA expire", soa("@", "mmm.ns.cloudflare.com.", "eee.cloudflare.com.", 2037190000, 10001, 2401, 604801, 3600)),
			tc("Modify SOA minttl", soa("@", "mmm.ns.cloudflare.com.", "eee.cloudflare.com.", 2037190000, 10001, 2401, 604801, 3601)),
		),

		testgroup("SRV", requires(providers.CanUseSRV),
			tc("SRV record", srv("_sip._tcp", 5, 6, 7, "foo.com.")),
			tc("Second SRV record, same prio", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 5, 60, 70, "foo2.com.")),
			tc("3 SRV", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 5, 60, 70, "foo2.com."), srv("_sip._tcp", 15, 65, 75, "foo3.com.")),
			tc("Delete one", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo3.com.")),
			tc("Change Target", srv("_sip._tcp", 5, 6, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
			tc("Change Priority", srv("_sip._tcp", 52, 6, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
			tc("Change Weight", srv("_sip._tcp", 52, 62, 7, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
			tc("Change Port", srv("_sip._tcp", 52, 62, 72, "foo.com."), srv("_sip._tcp", 15, 65, 75, "foo4.com.")),
			clear(),
			tc("Null Target", srv("_sip._tcp", 15, 65, 75, ".")),
		),

		testgroup("SSHFP",
			requires(providers.CanUseSSHFP),
			tc("SSHFP record",
				sshfp("@", 1, 1, "66c7d5540b7d75a1fb4c84febfa178ad99bdd67c")),
			tc("SSHFP change algorithm",
				sshfp("@", 2, 1, "66c7d5540b7d75a1fb4c84febfa178ad99bdd67c")),
			tc("SSHFP change fingerprint and type",
				sshfp("@", 2, 2, "745a635bc46a397a5c4f21d437483005bcc40d7511ff15fbfafe913a081559bc")),
		),

		testgroup("TLSA",
			requires(providers.CanUseTLSA),
			tc("TLSA record", tlsa("_443._tcp", 3, 1, 1, sha256hash)),
			tc("TLSA change usage", tlsa("_443._tcp", 2, 1, 1, sha256hash)),
			tc("TLSA change selector", tlsa("_443._tcp", 2, 0, 1, sha256hash)),
			tc("TLSA change matchingtype", tlsa("_443._tcp", 2, 0, 2, sha512hash)),
			tc("TLSA change certificate", tlsa("_443._tcp", 2, 0, 2, reversedSha512)),
		),

		testgroup("DS",
			requires(providers.CanUseDS),
			// Use a valid digest value here.  Some providers verify that a valid digest is in use.  See RFC 4034 and
			// https://www.iana.org/assignments/dns-sec-alg-numbers/dns-sec-alg-numbers.xhtml
			// https://www.iana.org/assignments/ds-rr-types/ds-rr-types.xhtml
			tc("DS create", ds("@", 1, 13, 1, "da39a3ee5e6b4b0d3255bfef95601890afd80709")),
			tc("DS change", ds("@", 8857, 8, 2, "4b9b6b073edd97feb5bc12dc4e1b32d2c6af7ae23a293936ceb87bb10494ec44")),
			tc("DS change f1", ds("@", 3, 8, 2, "4b9b6b073edd97feb5bc12dc4e1b32d2c6af7ae23a293936ceb87bb10494ec44")),
			tc("DS change f2", ds("@", 3, 13, 2, "4b9b6b073edd97feb5bc12dc4e1b32d2c6af7ae23a293936ceb87bb10494ec44")),
			tc("DS change f3+4", ds("@", 3, 13, 1, "da39a3ee5e6b4b0d3255bfef95601890afd80709")),
			tc("DS delete 1, create child", ds("another-child", 44, 13, 2, "4b9b6b073edd97feb5bc12dc4e1b32d2c6af7ae23a293936ceb87bb10494ec44")),
			tc("add 2 more DS",
				ds("another-child", 44, 13, 2, "4b9b6b073edd97feb5bc12dc4e1b32d2c6af7ae23a293936ceb87bb10494ec44"),
				ds("another-child", 1501, 13, 1, "ee02c885b5b4ed64899f2d43eb2b8e6619bdb50c"),
				ds("another-child", 1502, 8, 2, "2fa14f53e6b15cac9ac77846c7be87862c2a7e9ec0c6cea319db939317f126ed"),
				ds("another-child", 65535, 13, 2, "2fa14f53e6b15cac9ac77846c7be87862c2a7e9ec0c6cea319db939317f126ed"),
			),
			// These are the same as below.
			tc("DSchild create", ds("child", 1, 13, 1, "da39a3ee5e6b4b0d3255bfef95601890afd80709")),
			tc("DSchild change", ds("child", 8857, 8, 2, "4b9b6b073edd97feb5bc12dc4e1b32d2c6af7ae23a293936ceb87bb10494ec44")),
			tc("DSchild change f1", ds("child", 3, 8, 2, "4b9b6b073edd97feb5bc12dc4e1b32d2c6af7ae23a293936ceb87bb10494ec44")),
			tc("DSchild change f2", ds("child", 3, 13, 2, "4b9b6b073edd97feb5bc12dc4e1b32d2c6af7ae23a293936ceb87bb10494ec44")),
			tc("DSchild change f3+4", ds("child", 3, 13, 1, "da39a3ee5e6b4b0d3255bfef95601890afd80709")),
			tc("DSchild delete 1, create child", ds("another-child", 44, 13, 2, "4b9b6b073edd97feb5bc12dc4e1b32d2c6af7ae23a293936ceb87bb10494ec44")),
		),

		testgroup("DS (children only)",
			requires(providers.CanUseDSForChildren), not("CLOUDNS", "CLOUDFLAREAPI"),
			// Use a valid digest value here.  Some providers verify that a valid digest is in use.  See RFC 4034 and
			// https://www.iana.org/assignments/dns-sec-alg-numbers/dns-sec-alg-numbers.xhtml
			// https://www.iana.org/assignments/ds-rr-types/ds-rr-types.xhtml
			tc("DSchild create", ds("child", 1, 14, 4, "417212fd1c8bc5896fefd8db58af824545e85b0d0546409366a30aef7269fae258173bd185fb262c86f3bb86fba04368")),
			tc("DSchild change", ds("child", 8857, 8, 2, "4b9b6b073edd97feb5bc12dc4e1b32d2c6af7ae23a293936ceb87bb10494ec44")),
			tc("DSchild change f1", ds("child", 3, 8, 2, "4b9b6b073edd97feb5bc12dc4e1b32d2c6af7ae23a293936ceb87bb10494ec44")),
			tc("DSchild change f2", ds("child", 3, 13, 2, "4b9b6b073edd97feb5bc12dc4e1b32d2c6af7ae23a293936ceb87bb10494ec44")),
			tc("DSchild change f3+4", ds("child", 3, 14, 4, "3115238f89e0bf5252d9718113b1b9fff854608d84be94eefb9210dc1cc0b4f3557342a27465cfacc42ef137ae9a5489")),
			tc("DSchild delete 1, create child", ds("another-child", 44, 13, 2, "4b9b6b073edd97feb5bc12dc4e1b32d2c6af7ae23a293936ceb87bb10494ec44")),
			tc("add 2 more DSchild",
				ds("another-child", 44, 13, 2, "4b9b6b073edd97feb5bc12dc4e1b32d2c6af7ae23a293936ceb87bb10494ec44"),
				ds("another-child", 1501, 14, 4, "109bb6b5b6d5547c1ce03c7a8bd7d8f80c1cb0957f50c4f7fda04692079917e4f9cad52b878f3d8234e1a170b154b72d"),
				ds("another-child", 1502, 8, 2, "2fa14f53e6b15cac9ac77846c7be87862c2a7e9ec0c6cea319db939317f126ed"),
				ds("another-child", 65535, 13, 2, "2fa14f53e6b15cac9ac77846c7be87862c2a7e9ec0c6cea319db939317f126ed"),
			),
		),

		testgroup("DS (children only) CLOUDNS",
			requires(providers.CanUseDSForChildren),
			only("CLOUDNS", "CLOUDFLAREAPI"),
			// Cloudns requires NS records before creating DS Record. Verify
			// they are done in the right order, even if they are listed in
			// the wrong order in dnsconfig.js.
			tc("create DS",
				// we test that provider correctly handles creating NS first by reversing the entries here
				ds("child", 35632, 13, 1, "1E07663FF507A40874B8605463DD41DE482079D6"),
				ns("child", "ns101.cloudns.net."),
			),
			tc("modify field 1",
				ds("child", 2075, 13, 1, "2706D12E256C8FDD9BFB45EFB25FE537E21A82F6"),
				ns("child", "ns101.cloudns.net."),
			),
			tc("modify field 3",
				ds("child", 2075, 13, 2, "3F7A1EAC8C813A0BEBD0C3B8AAB387E31945EA0CD5E1D84A2E8E27674566C156"),
				ns("child", "ns101.cloudns.net."),
			),
			tc("modify field 2+3",
				ds("child", 2159, 1, 4, "F50BEFEA333EE2901D72D31A08E1A3CD3F7E943FF4B38CF7C8AD92807F5302F76FB0B419182C0F47FFC71CBCB6EF4BD4"),
				ns("child", "ns101.cloudns.net."),
			),
			tc("modify field 2",
				ds("child", 63909, 3, 4, "EEC7FA02E6788DA889B2CE41D43D92F948AB126EDCF83B7037E73CE9531C8E7E45653ABBAA76C2D6E42F98316EDE599B"),
				ns("child", "ns101.cloudns.net."),
			),
			//tc("modify field 2", ds("child", 65535, 254, 4, "0123456789ABCDEF")),
			tc("delete 1, create 1",
				ds("another-child", 35632, 13, 4, "F5F32ABCA6B01AA7A9963012F90B7C8523A1D946185A3AD70B67F3C9F18E7312FA9DD6AB2F7D8382F789213DB173D429"),
				ns("another-child", "ns101.cloudns.net."),
			),
			tc("add 2 more DS",
				ds("another-child", 35632, 13, 4, "F5F32ABCA6B01AA7A9963012F90B7C8523A1D946185A3AD70B67F3C9F18E7312FA9DD6AB2F7D8382F789213DB173D429"),
				ds("another-child", 2159, 1, 4, "F50BEFEA333EE2901D72D31A08E1A3CD3F7E943FF4B38CF7C8AD92807F5302F76FB0B419182C0F47FFC71CBCB6EF4BD4"),
				ds("another-child", 63909, 3, 4, "EEC7FA02E6788DA889B2CE41D43D92F948AB126EDCF83B7037E73CE9531C8E7E45653ABBAA76C2D6E42F98316EDE599B"),
				ns("another-child", "ns101.cloudns.net."),
			),
			// in CLouDNS  we must delete DS Record before deleting NS record
			// should no longer be necessary, provider should handle order correctly
			//tc("delete all DS",
			//	ns("another-child", "ns101.cloudns.net."),
			//),
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

		// AZURE features

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

		// ROUTE43 features

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

		// CLOUDFLAREAPI features

		testgroup("CF_REDIRECT",
			only("CLOUDFLAREAPI"),
			tc("redir", cfRedir("cnn.**current-domain-no-trailing**/*", "https://www.cnn.com/$1")),
			tc("change", cfRedir("cnn.**current-domain-no-trailing**/*", "https://change.cnn.com/$1")),
			tc("changelabel", cfRedir("cable.**current-domain-no-trailing**/*", "https://change.cnn.com/$1")),

			// Removed these for speed.  They were testing if order matters,
			// which it doesn't seem to.  Re-add if needed.
			//clear(),
			//tc("multipleA",
			//	cfRedir("cnn.**current-domain-no-trailing**/*", "https://www.cnn.com/$1"),
			//	cfRedir("msnbc.**current-domain-no-trailing**/*", "https://msnbc.cnn.com/$1"),
			//),
			//clear(),
			//tc("multipleB",
			//	cfRedir("msnbc.**current-domain-no-trailing**/*", "https://msnbc.cnn.com/$1"),
			//	cfRedir("cnn.**current-domain-no-trailing**/*", "https://www.cnn.com/$1"),
			//),
			//tc("change1",
			//	cfRedir("msnbc.**current-domain-no-trailing**/*", "https://msnbc.cnn.com/$1"),
			//	cfRedir("cnn.**current-domain-no-trailing**/*", "https://change.cnn.com/$1"),
			//),
			//tc("change1",
			//	cfRedir("msnbc.**current-domain-no-trailing**/*", "https://msnbc.cnn.com/$1"),
			//	cfRedir("cablenews.**current-domain-no-trailing**/*", "https://change.cnn.com/$1"),
			//),

			// TODO(tlim): Fix this test case. It is currently failing.
			//clear(),
			//tc("multiple3",
			//	cfRedir("msnbc.**current-domain-no-trailing**/*", "https://msnbc.cnn.com/$1"),
			//	cfRedir("cnn.**current-domain-no-trailing**/*", "https://www.cnn.com/$1"),
			//	cfRedir("nytimes.**current-domain-no-trailing**/*", "https://www.nytimes.com/$1"),
			//),

			// Repeat the above using CF_TEMP_REDIR instead
			clear(),
			tc("tempredir", cfRedirTemp("cnn.**current-domain-no-trailing**/*", "https://www.cnn.com/$1")),
			tc("tempchange", cfRedirTemp("cnn.**current-domain-no-trailing**/*", "https://change.cnn.com/$1")),
			tc("tempchangelabel", cfRedirTemp("cable.**current-domain-no-trailing**/*", "https://change.cnn.com/$1")),
			clear(),
			tc("tempmultipleA",
				cfRedirTemp("cnn.**current-domain-no-trailing**/*", "https://www.cnn.com/$1"),
				cfRedirTemp("msnbc.**current-domain-no-trailing**/*", "https://msnbc.cnn.com/$1"),
			),
			clear(),
			tc("tempmultipleB",
				cfRedirTemp("msnbc.**current-domain-no-trailing**/*", "https://msnbc.cnn.com/$1"),
				cfRedirTemp("cnn.**current-domain-no-trailing**/*", "https://www.cnn.com/$1"),
			),
			tc("tempchange1",
				cfRedirTemp("msnbc.**current-domain-no-trailing**/*", "https://msnbc.cnn.com/$1"),
				cfRedirTemp("cnn.**current-domain-no-trailing**/*", "https://change.cnn.com/$1"),
			),
			tc("tempchange1",
				cfRedirTemp("msnbc.**current-domain-no-trailing**/*", "https://msnbc.cnn.com/$1"),
				cfRedirTemp("cablenews.**current-domain-no-trailing**/*", "https://change.cnn.com/$1"),
			),
			// TODO(tlim): Fix this test case:
			//clear(),
			//tc("tempmultiple3",
			//	cfRedirTemp("msnbc.**current-domain-no-trailing**/*", "https://msnbc.cnn.com/$1"),
			//	cfRedirTemp("cnn.**current-domain-no-trailing**/*", "https://www.cnn.com/$1"),
			//	cfRedirTemp("nytimes.**current-domain-no-trailing**/*", "https://www.nytimes.com/$1"),
			//),
		),

		testgroup("CF_PROXY",
			only("CLOUDFLAREAPI"),
			tc("proxyon", cfProxyA("proxyme", "1.2.3.4", "on")),
			tc("proxychangetarget", cfProxyA("proxyme", "1.2.3.5", "on")),
			tc("proxychangeproxy", cfProxyA("proxyme", "1.2.3.5", "off")),
			clear(),
			tc("proxycname", cfProxyCNAME("anewproxy", "example.com.", "on")),
			tc("proxycnamechange", cfProxyCNAME("anewproxy", "example.com.", "off")),
			clear(),
		),

		testgroup("CF_WORKER_ROUTE",
			only("CLOUDFLAREAPI"),
			alltrue(*enableCFWorkers),
			// TODO(fdcastel): Add worker scripts via api call before test execution
			tc("simple", cfWorkerRoute("cnn.**current-domain-no-trailing**/*", "dnscontrol_integrationtest_cnn")),
			tc("changeScript", cfWorkerRoute("cnn.**current-domain-no-trailing**/*", "dnscontrol_integrationtest_msnbc")),
			tc("changePattern", cfWorkerRoute("cable.**current-domain-no-trailing**/*", "dnscontrol_integrationtest_msnbc")),
			clear(),
			tc("createMultiple",
				cfWorkerRoute("cnn.**current-domain-no-trailing**/*", "dnscontrol_integrationtest_cnn"),
				cfWorkerRoute("msnbc.**current-domain-no-trailing**/*", "dnscontrol_integrationtest_msnbc"),
			),
			tc("addOne",
				cfWorkerRoute("msnbc.**current-domain-no-trailing**/*", "dnscontrol_integrationtest_msnbc"),
				cfWorkerRoute("cnn.**current-domain-no-trailing**/*", "dnscontrol_integrationtest_cnn"),
				cfWorkerRoute("api.**current-domain-no-trailing**/cnn/*", "dnscontrol_integrationtest_cnn"),
			),
			tc("changeOne",
				cfWorkerRoute("msn.**current-domain-no-trailing**/*", "dnscontrol_integrationtest_msnbc"),
				cfWorkerRoute("cnn.**current-domain-no-trailing**/*", "dnscontrol_integrationtest_cnn"),
				cfWorkerRoute("api.**current-domain-no-trailing**/cnn/*", "dnscontrol_integrationtest_cnn"),
			),
			tc("deleteOne",
				cfWorkerRoute("msn.**current-domain-no-trailing**/*", "dnscontrol_integrationtest_msnbc"),
				cfWorkerRoute("api.**current-domain-no-trailing**/cnn/*", "dnscontrol_integrationtest_cnn"),
			),
		),

		// IGNORE* features

		testgroup("IGNORE_NAME function",
			tc("Create some records",
				txt("foo", "simple"),
				a("foo", "1.2.3.4"),
			),
			tc("Add a new record - ignoring foo",
				a("bar", "1.2.3.4"),
				ignoreName("foo"),
			),
			clear(),
			tc("Create some records",
				txt("bar.foo", "simple"),
				a("bar.foo", "1.2.3.4"),
			),
			tc("Add a new record - ignoring *.foo",
				a("bar", "1.2.3.4"),
				ignoreName("*.foo"),
			),
		),

		testgroup("IGNORE_NAME apex",
			tc("Create some records",
				txt("@", "simple"),
				a("@", "1.2.3.4"),
				txt("bar", "stringbar"),
				a("bar", "2.4.6.8"),
			),
			tc("Add a new record - ignoring apex",
				txt("bar", "stringbar"),
				a("bar", "2.4.6.8"),
				a("added", "4.6.8.9"),
				ignoreName("@"),
			),
		),

		testgroup("IGNORE_TARGET function",
			tc("Create some records",
				cname("foo", "test.foo.com."),
				cname("bar", "test.bar.com."),
			),
			tc("Add a new record - ignoring test.foo.com.",
				cname("bar", "bar.foo.com."),
				ignoreTarget("test.foo.com.", "CNAME"),
			),
			clear(),
			tc("Create some records",
				cname("bar.foo", "a.b.foo.com."),
				a("test.foo", "1.2.3.4"),
			),
			tc("Add a new record - ignoring **.foo.com. targets",
				a("bar", "1.2.3.4"),
				ignoreTarget("**.foo.com.", "CNAME"),
			),
		),
		// NB(tlim): We don't have a test for IGNORE_TARGET at the apex
		// because IGNORE_TARGET only works on CNAMEs and you can't have a
		// CNAME at the apex.  If we extend IGNORE_TARGET to support other
		// types of records, we should add a test at the apex.

	}

	return tests
}
