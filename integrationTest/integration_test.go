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

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/credsfile"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/pkg/nameservers"
	"github.com/StackExchange/dnscontrol/v4/pkg/zonerecs"
	"github.com/StackExchange/dnscontrol/v4/providers"
	_ "github.com/StackExchange/dnscontrol/v4/providers/_all"
	"github.com/StackExchange/dnscontrol/v4/providers/cloudflare"
	"github.com/miekg/dns/dnsutil"
)

var providerToRun = flag.String("provider", "", "Provider to run")
var startIdx = flag.Int("start", -1, "Test number to begin with")
var endIdx = flag.Int("end", -1, "Test index to stop after")
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
	dc.UpdateSplitHorizonNames()

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

	// Does this test require "diff2"?
	if f.diff2only && !diff2.EnableDiff2 {
		return fmt.Errorf("test for diff2 only")
	}

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
		if *providerToRun == "AXFRDDNS" {
			// Bind will refuse a DDNS update when the resulting zone
			// contains a NS record without an associated address
			// records (A or AAAA)
			dom.Records = append(dom.Records, a("ns."+domainName+".", "9.8.7.6"))
		}
		dom.IgnoredNames = tst.IgnoredNames
		dom.IgnoredTargets = tst.IgnoredTargets
		dom.Unmanaged = tst.Unmanaged
		dom.UnmanagedUnsafe = tst.UnmanagedUnsafe
		models.PostProcessRecords(dom.Records)
		dom2, _ := dom.Copy()

		if err := providers.AuditRecords(*providerToRun, dom.Records); err != nil {
			t.Skipf("***SKIPPED(PROVIDER DOES NOT SUPPORT '%s' ::%q)", err, desc)
			return
		}

		// get and run corrections for first time
		_, corrections, err := zonerecs.CorrectZoneRecords(prv, dom)
		if err != nil {
			t.Fatal(fmt.Errorf("runTests: %w", err))
		}
		if tst.Changeless {
			if count := len(corrections); count != 0 {
				t.Logf("Expected 0 corrections on FIRST run, but found %d.", count)
				for i, c := range corrections {
					t.Logf("UNEXPECTED #%d: %s", i, c.Msg)
				}
				t.FailNow()
			}
		} else if (len(corrections) == 0 && expectChanges) && (tst.Desc != "Empty") {
			t.Fatalf("Expected changes, but got none")
		}
		for _, c := range corrections {
			if *verbose {
				t.Log("\n" + c.Msg)
			}
			if c.F != nil { // F == nil if there is just a msg, no action.
				err = c.F()
				if err != nil {
					t.Fatal(err)
				}
			}
		}

		// If we just emptied out the zone, no need for a second pass.
		if len(tst.Records) == 0 {
			return
		}

		// run a second time and expect zero corrections
		_, corrections, err = zonerecs.CorrectZoneRecords(prv, dom2)
		if err != nil {
			t.Fatal(err)
		}
		if count := len(corrections); count != 0 {
			t.Logf("Expected 0 corrections on second run, but found %d.", count)
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
	if firstGroup == -1 {
		firstGroup = 0
	}
	lastGroup := *endIdx
	if lastGroup == -1 {
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

			// TODO(tlim): This is the old version. It skipped the remaining tc() statements if one failed.
			// The new code continues to test the remaining tc() statements.  Keeping this as a comment
			// in case we ever want to do something similar.
			// https://github.com/StackExchange/dnscontrol/pull/2252#issuecomment-1492204409
			//      makeChanges(t, prv, dc, tst, fmt.Sprintf("%02d:%s", gIdx, group.Desc), true, origConfig)
			//      if t.Failed() {
			//        break
			//      }
			if ok := makeChanges(t, prv, dc, tst, fmt.Sprintf("%02d:%s", gIdx, group.Desc), true, origConfig); !ok {
				break
			}

		}

		// Remove all records so next group starts with a clean slate.
		makeChanges(t, prv, dc, tc("Empty"), "Post cleanup", true, nil)

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

		rs, cs, err := zonerecs.CorrectZoneRecords(p, dom)
		if err != nil {
			t.Fatal(err)
		}
		for i, c := range rs {
			t.Logf("INFO#%d:\n%s", i+1, c.Msg)
		}
		for i, c := range cs {
			t.Logf("#%d:\n%s", i+1, c.Msg)
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
	rs, cs, err := zonerecs.CorrectZoneRecords(p, dc)
	if err != nil {
		t.Fatal(err)
	}
	if count := len(cs); count != 0 {
		t.Logf("Expect no corrections on second run, but found %d.", count)
		for i, c := range rs {
			t.Logf("INFO#%d:\n%s", i+1, c.Msg)
		}
		for i, c := range cs {
			t.Logf("#%d: %s", i+1, c.Msg)
		}
		t.FailNow()
	}
}

func TestNameserverDots(t *testing.T) {
	// Issue https://github.com/StackExchange/dnscontrol/issues/491
	// If this fails, the provider's GetNameservers() function uses
	// models.ToNameserversStripTD() instead of models.ToNameservers()
	// or vise-versa.

	// Setup:
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

	t.Run("No trailing dot in nameserver", func(t *testing.T) {
		for _, nameserver := range dc.Nameservers {
			//fmt.Printf("DEBUG: nameserver.Name = %q\n", nameserver.Name)
			if strings.HasSuffix(nameserver.Name, ".") {
				t.Errorf("Provider returned nameserver with trailing dot: %q", nameserver)
			}
		}
	})
}

type TestGroup struct {
	Desc      string
	required  []providers.Capability
	only      []string
	not       []string
	trueflags []bool
	tests     []*TestCase
	diff2only bool
}

type TestCase struct {
	Desc            string
	Records         []*models.RecordConfig
	IgnoredNames    []*models.IgnoreName
	IgnoredTargets  []*models.IgnoreTarget
	Unmanaged       []*models.UnmanagedConfig
	UnmanagedUnsafe bool // DISABLE_IGNORE_SAFETY_CHECK
	Changeless      bool // set to true if any changes would be an error
}

// ExpectNoChanges indicates that no changes is not an error, it is a requirement.
func (tc *TestCase) ExpectNoChanges() *TestCase {
	tc.Changeless = true
	return tc
}

// UnsafeIgnore is the equivalent of DISABLE_IGNORE_SAFETY_CHECK
func (tc *TestCase) UnsafeIgnore() *TestCase {
	tc.UnmanagedUnsafe = true
	return tc
}

func (tg *TestGroup) Diff2Only() *TestGroup {
	tg.diff2only = true
	return tg
}

func SetLabel(r *models.RecordConfig, label, domain string) {
	r.Name = label
	r.NameFQDN = dnsutil.AddOrigin(label, "**current-domain**")
}

func a(name, target string) *models.RecordConfig {
	return makeRec(name, target, "A")
}

func alias(name, target string) *models.RecordConfig {
	return makeRec(name, target, "ALIAS")
}

func azureAlias(name, aliasType, target string) *models.RecordConfig {
	r := makeRec(name, target, "AZURE_ALIAS")
	r.AzureAlias = map[string]string{
		"type": aliasType,
	}
	return r
}

func caa(name string, tag string, flag uint8, target string) *models.RecordConfig {
	r := makeRec(name, target, "CAA")
	r.SetTargetCAA(flag, tag, target)
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

func cname(name, target string) *models.RecordConfig {
	return makeRec(name, target, "CNAME")
}

func ds(name string, keyTag uint16, algorithm, digestType uint8, digest string) *models.RecordConfig {
	r := makeRec(name, "", "DS")
	r.SetTargetDS(keyTag, algorithm, digestType, digest)
	return r
}

func ignoreName(labelSpec string) *models.RecordConfig {
	r := &models.RecordConfig{
		Type:     "IGNORE_NAME",
		Metadata: map[string]string{},
	}
	// diff1
	SetLabel(r, labelSpec, "**current-domain**")
	// diff2
	r.Metadata["ignore_LabelPattern"] = labelSpec
	return r
}

func ignoreTarget(targetSpec string, typeSpec string) *models.RecordConfig {
	r := &models.RecordConfig{
		Type:     "IGNORE_TARGET",
		Metadata: map[string]string{},
	}
	// diff1
	r.SetTarget(typeSpec)
	SetLabel(r, targetSpec, "**current-domain**")
	// diff2
	r.Metadata["ignore_RTypePattern"] = typeSpec
	r.Metadata["ignore_TargetPattern"] = typeSpec
	return r
}

func ignore(labelSpec string, typeSpec string, targetSpec string) *models.RecordConfig {
	r := &models.RecordConfig{
		Type:     "IGNORE",
		Metadata: map[string]string{},
	}
	if r.Metadata == nil {
		r.Metadata = map[string]string{}
	}
	r.Metadata["ignore_LabelPattern"] = labelSpec
	r.Metadata["ignore_RTypePattern"] = typeSpec
	r.Metadata["ignore_TargetPattern"] = targetSpec
	return r
}

func loc(name string, d1 uint8, m1 uint8, s1 float32, ns string,
	d2 uint8, m2 uint8, s2 float32, ew string, al int32, sz float32, hp float32, vp float32) *models.RecordConfig {
	r := makeRec(name, "", "LOC")
	r.SetLOCParams(d1, m1, s1, ns, d2, m2, s2, ew, al, sz, hp, vp)
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

func manyA(namePattern, target string, n int) []*models.RecordConfig {
	recs := []*models.RecordConfig{}
	for i := 0; i < n; i++ {
		recs = append(recs, makeRec(fmt.Sprintf(namePattern, i), target, "A"))
	}
	return recs
}

func mx(name string, prio uint16, target string) *models.RecordConfig {
	r := makeRec(name, target, "MX")
	r.MxPreference = prio
	return r
}

func ns(name, target string) *models.RecordConfig {
	return makeRec(name, target, "NS")
}

func naptr(name string, order uint16, preference uint16, flags string, service string, regexp string, target string) *models.RecordConfig {
	r := makeRec(name, target, "NAPTR")
	r.SetTargetNAPTR(order, preference, flags, service, regexp, target)
	return r
}

func ptr(name, target string) *models.RecordConfig {
	return makeRec(name, target, "PTR")
}

func r53alias(name, aliasType, target string) *models.RecordConfig {
	r := makeRec(name, target, "R53_ALIAS")
	r.R53Alias = map[string]string{
		"type": aliasType,
	}
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
	var unmanagedItems []*models.UnmanagedConfig
	for _, r := range recs {
		switch r.Type {
		case "IGNORE":
			// diff1:
			ignoredNames = append(ignoredNames, &models.IgnoreName{
				Pattern: r.Metadata["ignore_LabelPattern"],
				Types:   r.Metadata["ignore_RTypePattern"],
			})
			// diff2:
			unmanagedItems = append(unmanagedItems, &models.UnmanagedConfig{
				LabelPattern:  r.Metadata["ignore_LabelPattern"],
				RTypePattern:  r.Metadata["ignore_RTypePattern"],
				TargetPattern: r.Metadata["ignore_TargetPattern"],
			})
			continue
		case "IGNORE_NAME":
			ignoredNames = append(ignoredNames, &models.IgnoreName{Pattern: r.GetLabel(), Types: r.GetTargetField()})
			unmanagedItems = append(unmanagedItems, &models.UnmanagedConfig{
				LabelPattern: r.GetLabel(),
				RTypePattern: r.GetTargetField(),
			})
			continue
		case "IGNORE_TARGET":
			ignoredTargets = append(ignoredTargets, &models.IgnoreTarget{
				Pattern: r.GetLabel(),
				Type:    r.GetTargetField(),
			})
			unmanagedItems = append(unmanagedItems, &models.UnmanagedConfig{
				RTypePattern:  r.GetTargetField(),
				TargetPattern: r.GetLabel(),
			})
		default:
			records = append(records, r)
		}
	}
	return &TestCase{
		Desc:           desc,
		Records:        records,
		IgnoredNames:   ignoredNames,
		IgnoredTargets: ignoredTargets,
		Unmanaged:      unmanagedItems,
	}
}

func txt(name, target string) *models.RecordConfig {
	r := makeRec(name, "", "TXT")
	r.SetTargetTXT(target)
	return r
}

// func (r *models.RecordConfig) ttl(t uint32) *models.RecordConfig {
func ttl(r *models.RecordConfig, t uint32) *models.RecordConfig {
	r.TTL = t
	return r
}

func tlsa(name string, usage, selector, matchingtype uint8, target string) *models.RecordConfig {
	r := makeRec(name, target, "TLSA")
	r.SetTargetTLSA(usage, selector, matchingtype, target)
	return r
}

func ns1Urlfwd(name, target string) *models.RecordConfig {
	return makeRec(name, target, "NS1_URLFWD")
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

		// START HERE

		// Narrative:  Hello friend!  Are you adding a new DNS provider to
		// DNSControl? That's awesome!  I'm here to help.
		//
		// As you write your code, these tests will help verify that your
		// code is correct and covers all the funny edge-cases that DNS
		// providers throw at us.
		//
		// If you follow these sections marked "Narrative", I'll lead you
		// through the tests. The tests start by testing very basic things
		// (are you talking to the API correctly) and then moves on to
		// more and more esoteric issues.  It's like a video game where
		// you have to solve all the levels but the game lets you skip
		// around as long as all the levels are completed eventually.  Some
		// of the levels you can mark "not relevant" for your provider.
		//
		// Oh wait. I'm getting ahead of myself.  How do you run these
		// tests?  That's documented here:
		// https://docs.dnscontrol.org/developer-info/integration-tests
		// You'll be running these tests a lot. I recommend you make a
		// script that sets the environment variables and runs the tests
		// to make it easy to run the tests.  However don't check that
		// file into a GIT repo... it contains API credentials that are
		// secret!

		///// Basic functionality (add/rename/change/delete).

		// Narrative:  Let's get started!  The first thing to do is to
		// make sure we can create an A record, change it, then delete it.
		// That's the basic Add/Change/Delete process.  Once these three
		// features work you know that your API calls and authentication
		// is working and we can do the most basic operations.

		testgroup("A",
			tc("Create A", a("testa", "1.1.1.1")),
			tc("Change A target", a("testa", "3.3.3.3")),
		),

		// Narrative: Congrats on getting those to work!  Now let's try
		// something a little more difficult.  Let's do that same test at
		// the apex of the domain.  This may "just work" for your
		// provider, or they might require something special like
		// referring to the apex as "@".

		// Same test, but at the apex of the domain.
		testgroup("Apex",
			tc("Create A", a("@", "2.2.2.2")),
			tc("Change A target", a("@", "4.4.4.4")),
		),

		// Narrative: Another edge-case is the wildcard record ("*").  In
		// theory this should "just work" but plenty of vendors require
		// some weird quoting or escaping. None of that should be required
		// but... sigh... they do it anyway.  Let's find out how badly
		// they screwed this up!

		// Same test, but do it with a wildcard.
		testgroup("Protocol-Wildcard",
			not("HEDNS"), // Not supported by dns.he.net due to abuse
			tc("Create wildcard", a("*", "3.3.3.3"), a("www", "5.5.5.5")),
			tc("Delete wildcard", a("www", "5.5.5.5")),
		),

		///// Test the basic DNS types

		// Narrative: That wasn't as hard as expected, eh?  Let's test the
		// other basic record types like AAAA, CNAME, MX and TXT.

		// AAAA: TODO(tlim) Add AAAA test.

		// CNAME

		testgroup("CNAME",
			tc("Create a CNAME", cname("testcname", "www.google.com.")),
			tc("Change CNAME target", cname("testcname", "www.yahoo.com.")),
		),

		// MX

		// Narrative: MX is the first record we're going to test with
		// multiple fields. All records have a target (A records have an
		// IP address, CNAMEs have a destination (called "the canonical
		// name" in the RFCs). MX records have a target (a hostname) but
		// also have a "Preference".  FunFact: The RFCs call this the
		// "preference" but most engineers refer to it as the "priority".
		// Now you know better.
		// Let's make sure your code creates and updates the preference
		// correctly!

		testgroup("MX",
			tc("Create MX", mx("testmx", 5, "foo.com.")),
			tc("Change MX target", mx("testmx", 5, "bar.com.")),
			tc("Change MX p", mx("testmx", 100, "bar.com.")),
		),

		// TXT

		// Narrative: TXT records can be very complex but we'll save those
		// tests for later. Let's just test a simple string.

		testgroup("TXT",
			tc("Create TXT", txt("testtxt", "simple")),
			tc("Change TXT target", txt("testtxt", "changed")),
		),

		// Test API edge-cases

		// Narrative: I'm proud of you for getting this far.  All the
		// basic types work!  Now let's verify your code handles some of
		// the more interesting ways that updates can happen.  For
		// example, let's try creating many records of the same or
		// different type at once.  Usually this "just works" but maybe
		// there's an off-by-one error lurking. Once these work we'll have
		// a new level of confidence in the code.

		testgroup("ManyAtOnce",
			tc("CreateManyAtLabel", a("www", "1.1.1.1"), a("www", "2.2.2.2"), a("www", "3.3.3.3")),
			clear(),
			tc("Create an A record", a("www", "1.1.1.1")),
			tc("Add at label1", a("www", "1.1.1.1"), a("www", "2.2.2.2")),
			tc("Add at label2", a("www", "1.1.1.1"), a("www", "2.2.2.2"), a("www", "3.3.3.3")),
		),

		testgroup("manyTypesAtOnce",
			tc("CreateManyTypesAtLabel", a("www", "1.1.1.1"), mx("testmx", 5, "foo.com."), mx("testmx", 100, "bar.com.")),
			clear(),
			tc("Create an A record", a("www", "1.1.1.1")),
			tc("Add Type At Label", a("www", "1.1.1.1"), mx("testmx", 5, "foo.com.")),
			tc("Add Type At Label", a("www", "1.1.1.1"), mx("testmx", 5, "foo.com."), mx("testmx", 100, "bar.com.")),
		),

		// Exercise TTL operations.

		// Narrative: TTLs are weird.  They deserve some special tests.
		// First we'll verify some simple cases but then we'll test the
		// weirdest edge-case we've ever seen.

		testgroup("Attl",
			not("LINODE"), // Linode does not support arbitrary TTLs: both are rounded up to 3600.
			tc("Create Arc", ttl(a("testa", "1.1.1.1"), 333)),
			tc("Change TTL", ttl(a("testa", "1.1.1.1"), 999)),
		),

		testgroup("TTL",
			not("NETCUP"), // NETCUP does not support TTLs.
			not("LINODE"), // Linode does not support arbitrary TTLs: 666 and 1000 are both rounded up to 3600.
			tc("Start", ttl(a("@", "8.8.8.8"), 666), a("www", "1.2.3.4"), a("www", "5.6.7.8")),
			tc("Change a ttl", ttl(a("@", "8.8.8.8"), 1000), a("www", "1.2.3.4"), a("www", "5.6.7.8")),
			tc("Change single target from set", ttl(a("@", "8.8.8.8"), 1000), a("www", "2.2.2.2"), a("www", "5.6.7.8")),
			tc("Change all ttls", ttl(a("@", "8.8.8.8"), 500), ttl(a("www", "2.2.2.2"), 400), ttl(a("www", "5.6.7.8"), 400)),
		),

		// Narrative: Did you see that `not("NETCUP")` code?  NETCUP just
		// plain doesn't support TTLs, so those tests just plain can't
		// ever work.  `not("NETCUP")` tells the test system to skip those
		// tests. There's also `only()` which runs a test only for certain
		// providers.  Those and more are documented above in the
		// "Filters" section, which is on line 664 as I write this.

		// Narrative: Ok, back to testing.  This next test is a strange
		// one. It's a strange situation that happens rarely.  You might
		// want to skip this and come back later, or ask for help on the
		// mailing list.

		// Test: At the start we have a single DNS record at a label.
		// Next we add an additional record at the same label AND change
		// the TTL of the existing record.
		testgroup("add to label and change orig ttl",
			tc("Setup", ttl(a("www", "5.6.7.8"), 400)),
			tc("Add at same label, new ttl", ttl(a("www", "5.6.7.8"), 700), ttl(a("www", "1.2.3.4"), 700)),
		),

		// Narrative: We're done with TTL tests now.  If you fixed a bug
		// in any of those tests give yourself a pat on the back. Finding
		// bugs is not bad or shameful... it's an opportunity to help the
		// world by fixing a problem!  If only we could fix all the
		// world's problems by editing code!
		//
		// Now let's look at one more edge-case: Can you change the type
		// of a record?  Some providers don't permit this and you have to
		// delete the old record and create a new record in its place.

		testgroup("TypeChange",
			// Test whether the provider properly handles a label changing
			// from one rtype to another.
			tc("Create A", a("foo", "1.2.3.4")),
			tc("Change to MX", mx("foo", 5, "mx.google.com.")),
			tc("Change back to A", a("foo", "4.5.6.7")),
		),

		// Narrative: That worked? Of course that worked. You're awesome.
		// Now let's make it even more difficult by involving CNAMEs.  If
		// there is a CNAME at a label, no other records can be at that
		// label. That means the order of updates is critical when
		// changing A->CNAME or CNAME->A.  pkg/diff2 should order the
		// changes properly for you. Let's verify that we got it right!

		testgroup("TypeChangeHard",
			tc("Create a CNAME", cname("foo", "google.com.")),
			tc("Change to A record", a("foo", "1.2.3.4")),
			tc("Change back to CNAME", cname("foo", "google2.com.")),
		),

		//// Test edge cases from various types.

		// Narrative: Every DNS record type has some weird edge-case that
		// you wouldn't expect. This is where we test those situations.
		// They're strange, but usually easy to fix or skip.
		//
		// Some of these are testing the provider more than your code.
		//
		// You can't fix your provider's code. That's why there is the
		// auditrecord.go system.  For example, if your provider doesn't
		// support MX records that point to "." (yes, that's a thing),
		// there's nothing you can do other than warn users that it isn't
		// supported.  We do this in the auditrecords.go file in each
		// provider. It contains "rejectif.` statements that detect
		// unsupported situations.  Some good examples are in
		// providers/cscglobal/auditrecords.go. Take a minute to read
		// that.

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

		//// TXT tests

		// Narrative: TXT records are weird. It's just text, right?  Sadly
		// "just text" means quotes and other funny characters that might
		// need special handling. In some cases providers ban certain
		// chars in the string.
		//
		// Let's test the weirdness we've found.  I wouldn't bother trying
		// too hard to fix these. Just skip them by updating
		// auditrecords.go for your provider.

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

			//clear(),
			//tc("a 255-byte TXT", txt("foo255", strings.Repeat("C", 255))),
			//clear(),
			//tc("a 256-byte TXT", txt("foo256", strings.Repeat("D", 256))),
			//clear(),
			//tc("a 512-byte TXT", txt("foo512", strings.Repeat("C", 512))),
			//clear(),
			//tc("a 513-byte TXT", txt("foo513", strings.Repeat("D", 513))),
			//clear(),

			//tc("TXT with 1 single-quote", txt("foosq", "quo'te")),
			//clear(),
			//tc("TXT with 1 backtick", txt("foobt", "blah`blah")),
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

		// Narrative: Congratulate yourself for getting this far.
		// Seriously.  Buy yourself a beer or other beverage.  Kick back.
		// Take a break.  Ok, break over!  Time for some more weird edge
		// cases.

		// DNSControl downcases all DNS labels. These tests make sure
		// that's all done correctly.
		testgroup("Case Sensitivity",
			// The decoys are required so that there is at least one actual
			// change in each tc.
			tc("Create CAPS", mx("BAR", 5, "BAR.com.")),
			tc("Downcase label", mx("bar", 5, "BAR.com."), a("decoy", "1.1.1.1")),
			tc("Downcase target", mx("bar", 5, "bar.com."), a("decoy", "2.2.2.2")),
			tc("Upcase both", mx("BAR", 5, "BAR.COM."), a("decoy", "3.3.3.3")),
		),

		// Make sure we can manipulate one DNS record when there is
		// another at the same label.
		testgroup("testByLabel",
			tc("initial",
				a("foo", "1.2.3.4"),
				a("foo", "2.3.4.5"),
			),
			tc("changeOne",
				a("foo", "1.2.3.4"),
				a("foo", "3.4.5.6"), // Change
			),
			tc("deleteOne",
				a("foo", "1.2.3.4"),
				//a("foo", "3.4.5.6"), // Delete
			),
			tc("addOne",
				a("foo", "1.2.3.4"),
				a("foo", "3.4.5.6"), // Add
			),
		),

		// Make sure we can manipulate one DNS record when there is
		// another at the same RecordSet.
		testgroup("testByRecordSet",
			tc("initial",
				a("bar", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				a("foo", "3.4.5.6"),
				mx("foo", 10, "foo.**current-domain**"),
				mx("foo", 20, "bar.**current-domain**"),
			),
			tc("changeOne",
				a("bar", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				a("foo", "8.8.8.8"), // Change
				mx("foo", 10, "foo.**current-domain**"),
				mx("foo", 20, "bar.**current-domain**"),
			),
			tc("deleteOne",
				a("bar", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				//a("foo", "8.8.8.8"),  // Delete
				mx("foo", 10, "foo.**current-domain**"),
				mx("foo", 20, "bar.**current-domain**"),
			),
			tc("addOne",
				a("bar", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				a("foo", "8.8.8.8"), // Add
				mx("foo", 10, "foo.**current-domain**"),
				mx("foo", 20, "bar.**current-domain**"),
			),
		),

		// Narrative: Here we test the IDNA (internationalization)
		// features.  But first a joke:
		// Q: What do you call someone that speaks 2 languages?
		// A: bilingual
		// Q: What do you call someone that speaks 3 languages?
		// A: trilingual
		// Q: What do you call someone that speaks 1 language?
		// A: American
		// Get it?  Well, that's why I'm not a stand-up comedian.
		// Anyway... let's make sure foreign languages work.

		testgroup("IDNA",
			not("SOFTLAYER"),
			// SOFTLAYER: fails at direct internationalization, punycode works, of course.
			tc("Internationalized name", a("ööö", "1.2.3.4")),
			tc("Change IDN", a("ööö", "2.2.2.2")),
			tc("Internationalized CNAME Target", cname("a", "ööö.com.")),
		),
		testgroup("IDNAs in CNAME targets",
			not("CLOUDFLAREAPI"),
			// LINODE: hostname validation does not allow the target domain TLD
			tc("IDN CNAME AND Target", cname("öoö", "ööö.企业.")),
		),

		// Narrative: Some providers send the list of DNS records one
		// "page" at a time. The data you get includes a flag that
		// indicates you to the request is incomplete and you need to
		// request the next page of data.  They don't realize that
		// computers have gigabytes of RAM and the largest DNS zone might
		// have kilobytes of records.  Unneeded complexity... sigh.
		//
		// Let's test to make sure we got the paging right. I always fear
		// off-by-one errors when I write this kind of code. Like... if a
		// get tells you it has returned a page that starts at record 0
		// and includes 100 records, should the next "get" request records
		// starting at 99 or 100 or 101?
		//
		// These tests can be VERY slow. That's why we use not() and
		// only() to skip these tests for providers that doesn't use
		// paging.

		testgroup("pager101",
			// Tests the paging code of providers.  Many providers page at 100.
			// Notes:
			//  - Gandi: page size is 100, therefore we test with 99, 100, and 101
			//  - DIGITALOCEAN: page size is 100 (default: 20)
			not(
				"AZURE_DNS",     // Removed because it is too slow
				"CLOUDFLAREAPI", // Infinite pagesize but due to slow speed, skipping.
				"DIGITALOCEAN",  // No paging. Why bother?
				"CSCGLOBAL",     // Doesn't page. Works fine.  Due to the slow API we skip.
				"GANDI_V5",      // Their API is so damn slow. We'll add it back as needed.
				"HEDNS",         // Doesn't page. Works fine.  Due to the slow API we skip.
				"LOOPIA",        // Their API is so damn slow. Plus, no paging.
				"MSDNS",         // No paging done. No need to test.
				"NAMEDOTCOM",    // Their API is so damn slow. We'll add it back as needed.
				"NS1",           // Free acct only allows 50 records, therefore we skip
				//"ROUTE53",       // Batches up changes in pages.
			),
			tc("99 records", manyA("rec%04d", "1.2.3.4", 99)...),
			tc("100 records", manyA("rec%04d", "1.2.3.4", 100)...),
			tc("101 records", manyA("rec%04d", "1.2.3.4", 101)...),
		),

		testgroup("pager601",
			only(
				//"AZURE_DNS",     // Removed because it is too slow
				//"CLOUDFLAREAPI", // Infinite pagesize but due to slow speed, skipping.
				//"CSCGLOBAL",     // Doesn't page. Works fine.  Due to the slow API we skip.
				//"GANDI_V5",      // Their API is so damn slow. We'll add it back as needed.
				//"MSDNS",         // No paging done. No need to test.
				"GCLOUD",
				"HEXONET",
				"ROUTE53", // Batches up changes in pages.
			),
			tc("601 records", manyA("rec%04d", "1.2.3.4", 600)...),
			tc("Update 601 records", manyA("rec%04d", "1.2.3.5", 600)...),
		),

		testgroup("pager1201",
			only(
				//"AKAMAIEDGEDNS", // No paging done. No need to test.
				//"AZURE_DNS",     // Currently failing. See https://github.com/StackExchange/dnscontrol/issues/770
				//"CLOUDFLAREAPI", // Fails with >1000 corrections. See https://github.com/StackExchange/dnscontrol/issues/1440
				//"CSCGLOBAL",     // Doesn't page. Works fine.  Due to the slow API we skip.
				//"GANDI_V5",      // Their API is so damn slow. We'll add it back as needed.
				//"HEDNS",         // No paging done. No need to test.
				//"MSDNS",         // No paging done. No need to test.
				"HEXONET",
				"HOSTINGDE", // Pages.
				"ROUTE53",   // Batches up changes in pages.
			),
			tc("1200 records", manyA("rec%04d", "1.2.3.4", 1200)...),
			tc("Update 1200 records", manyA("rec%04d", "1.2.3.5", 1200)...),
		),

		//// CanUse* types:

		// Narrative: Many DNS record types are optional.  If the provider
		// supports them, there's a CanUse* variable that flags that
		// feature.  Here we test those.  Each of these should (1) create
		// the record, (2) test changing additional fields one at a time,
		// maybe 2 at a time, (3) delete the record. If you can do those 3
		// things, we're pretty sure you've implemented it correctly.

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

		// LOCation records. // No.47
		testgroup("LOC",
			requires(providers.CanUseLOC),
			//42 21 54     N  71 06  18     W -24m 30m
			tc("Single LOC record", loc("@", 42, 21, 54, "N", 71, 6, 18, "W", -24, 30, 0, 0)),
			//42 21 54     N  71 06  18     W -24m 30m
			tc("Update single LOC record", loc("@", 42, 21, 54, "N", 71, 6, 18, "W", -24, 30, 10, 0)),
			tc("Multiple LOC records-create a-d modify apex", //create a-d, modify @
				//42 21 54     N  71 06  18     W -24m 30m
				loc("@", 42, 21, 54, "N", 71, 6, 18, "W", -24, 30, 0, 0),
				//42 21 43.952 N  71 5   6.344  W -24m 1m 200m
				loc("a", 42, 21, 43.952, "N", 71, 5, 6.344, "W", -24, 1, 200, 10),
				//52 14 05     N  00 08  50     E 10m
				loc("b", 52, 14, 5, "N", 0, 8, 50, "E", 10, 0, 0, 0),
				//32  7 19     S 116  2  25     E 10m
				loc("c", 32, 7, 19, "S", 116, 2, 25, "E", 10, 0, 0, 0),
				//42 21 28.764 N  71 00  51.617 W -44m 2000m
				loc("d", 42, 21, 28.764, "N", 71, 0, 51.617, "W", -44, 2000, 0, 0),
			),
		),

		// Narrative: NAPTR records are used by IP telephony ("SIP")
		// systems. NAPTR records are rarely used, but if you use them
		// you'll want to use DNSControl because editing them is a pain.
		// If you want a fun read, check this out:
		// https://www.devever.net/~hl/sip-victory

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
		testgroup("PTR",
			requires(providers.CanUsePTR),
			not("CLOUDNS"),
			tc("Create PTR record", ptr("4", "foo.com.")),
			tc("Modify PTR record", ptr("4", "bar.com.")),
		),

		// Narrative: SOA records are ignored by most DNS providers. They
		// auto-generate the values and ignore your SOA data. Don't
		// implement the SOA record unless your provide can not work
		// without them, like BIND.

		// SOA
		testgroup("SOA",
			requires(providers.CanUseSOA),
			clear(), // Extra clear required or only the first run passes.
			tc("Create SOA record", soa("@", "kim.ns.cloudflare.com.", "dns.cloudflare.com.", 2037190000, 10000, 2400, 604800, 3600)),
			tc("Modify SOA ns    ", soa("@", "mmm.ns.cloudflare.com.", "dns.cloudflare.com.", 2037190000, 10000, 2400, 604800, 3600)),
			tc("Modify SOA mbox  ", soa("@", "mmm.ns.cloudflare.com.", "eee.cloudflare.com.", 2037190000, 10000, 2400, 604800, 3600)),
			tc("Modify SOA refres", soa("@", "mmm.ns.cloudflare.com.", "eee.cloudflare.com.", 2037190000, 10001, 2400, 604800, 3600)),
			tc("Modify SOA retry ", soa("@", "mmm.ns.cloudflare.com.", "eee.cloudflare.com.", 2037190000, 10001, 2401, 604800, 3600)),
			tc("Modify SOA expire", soa("@", "mmm.ns.cloudflare.com.", "eee.cloudflare.com.", 2037190000, 10001, 2401, 604801, 3600)),
			tc("Modify SOA minttl", soa("@", "mmm.ns.cloudflare.com.", "eee.cloudflare.com.", 2037190000, 10001, 2401, 604801, 3601)),
		),

		testgroup("SRV",
			requires(providers.CanUseSRV),
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

		// https://github.com/StackExchange/dnscontrol/issues/2066
		testgroup("SRV",
			requires(providers.CanUseSRV),
			tc("Create SRV333", ttl(srv("_sip._tcp", 5, 6, 7, "foo.com."), 333)),
			tc("Change TTL999", ttl(srv("_sip._tcp", 5, 6, 7, "foo.com."), 999)),
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
			requires(providers.CanUseDSForChildren),
			not("CLOUDNS", "CLOUDFLAREAPI"),
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

		//// Vendor-specific record types

		// Narrative: DNSControl supports DNS records that don't exist!
		// Well, they exist for particular vendors.  Let's test each of
		// them here. If you are writing a new provider, I have some good
		// news: These don't apply to you!

		testgroup("ALIAS",
			requires(providers.CanUseAlias),
			tc("ALIAS at root", alias("@", "foo.com.")),
			tc("change it", alias("@", "foo2.com.")),
			tc("ALIAS at subdomain", alias("test", "foo.com.")),
			tc("change it", alias("test", "foo2.com.")),
		),

		// AZURE features

		testgroup("AZURE_ALIAS_A",
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
			tc("change aliasA",
				a("foo.a", "1.2.3.4"),
				a("quux.a", "2.3.4.5"),
				azureAlias("bar.a", "A", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain-no-trailing**/A/quux.a"),
			),
			tc("change backA",
				a("foo.a", "1.2.3.4"),
				a("quux.a", "2.3.4.5"),
				azureAlias("bar.a", "A", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain-no-trailing**/A/foo.a"),
			),
		),

		testgroup("AZURE_ALIAS_CNAME",
			requires(providers.CanUseAzureAlias),
			tc("create dependent CNAME records",
				cname("foo.cname", "google.com."),
				cname("quux.cname", "google2.com."),
			),
			tc("ALIAS to CNAME record in same zone",
				cname("foo.cname", "google.com."),
				cname("quux.cname", "google2.com."),
				azureAlias("bar.cname", "CNAME", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain-no-trailing**/CNAME/foo.cname"),
			),
			tc("change aliasCNAME",
				cname("foo.cname", "google.com."),
				cname("quux.cname", "google2.com."),
				azureAlias("bar.cname", "CNAME", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain-no-trailing**/CNAME/quux.cname"),
			),
			tc("change backCNAME",
				cname("foo.cname", "google.com."),
				cname("quux.cname", "google2.com."),
				azureAlias("bar.cname", "CNAME", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain-no-trailing**/CNAME/foo.cname"),
			),
		),

		// ROUTE53 features

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
		),

		testgroup("R53_ALIAS_CNAME",
			requires(providers.CanUseRoute53Alias),
			tc("create alias+cname in one step",
				r53alias("dev-system", "CNAME", "dev-system18.**current-domain**"),
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
			),
		),

		testgroup("R53_ALIAS_Loop",
			// This will always be skipped because rejectifTargetEqualsLabel
			// will always flag it as not permitted.
			// See https://github.com/StackExchange/dnscontrol/issues/2107
			requires(providers.CanUseRoute53Alias),
			tc("loop should fail",
				r53alias("test-islandora", "CNAME", "test-islandora.**current-domain**"),
			),
		),

		// Bug https://github.com/StackExchange/dnscontrol/issues/2285
		testgroup("R53_alias pre-existing",
			requires(providers.CanUseRoute53Alias),
			tc("Create some records",
				r53alias("dev-system", "CNAME", "dev-system18.**current-domain**"),
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
			),
			tc("Add a new record - ignoring foo",
				a("bar", "1.2.3.4"),
				ignoreName("dev-system*"),
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
			tc("proxychangeonoff", cfProxyA("proxyme", "1.2.3.5", "off")),
			tc("proxychangeoffon", cfProxyA("proxyme", "1.2.3.5", "on")),
			clear(),
			tc("proxycname", cfProxyCNAME("anewproxy", "example.com.", "on")),
			tc("proxycnamechange", cfProxyCNAME("anewproxy", "example.com.", "off")),
			tc("proxycnameoffon", cfProxyCNAME("anewproxy", "example.com.", "on")),
			tc("proxycnameonoff", cfProxyCNAME("anewproxy", "example.com.", "off")),
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

		// NS1 features

		testgroup("NS1_URLFWD tests",
			only("NS1"),
			tc("Add a urlfwd", ns1Urlfwd("urlfwd1", "/ http://example.com 302 2 0")),
			tc("Update a urlfwd", ns1Urlfwd("urlfwd1", "/ http://example.org 301 2 0")),
		),

		//// IGNORE* features

		// Narrative: You're basically done now. These remaining tests
		// exercise the NO_PURGE and IGNORE* features.  These are handled
		// by the pkg/diff2 module. If they work for any provider, they
		// should work for all providers.  However we're going to test
		// them anyway because one never knows.  Ready?  Let's go!

		testgroup("IGNORE main",
			tc("Create some records",
				txt("foo", "simple"),
				a("foo", "1.2.3.4"),
				a("bar", "5.5.5.5"),
			),
			tc("ignore label=foo",
				a("bar", "5.5.5.5"),
				ignore("foo", "", ""),
			).ExpectNoChanges(),
			tc("ignore type=txt",
				a("foo", "1.2.3.4"),
				a("bar", "5.5.5.5"),
				ignore("", "TXT", ""),
			).ExpectNoChanges(),
			tc("ignore target=1.2.3.4",
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				ignore("", "", "1.2.3.4"),
			).ExpectNoChanges(),
			tc("ignore manytypes",
				ignore("", "A,TXT", ""),
			).ExpectNoChanges(),
		).Diff2Only(),

		testgroup("IGNORE apex",
			tc("Create some records",
				txt("@", "simple"),
				a("@", "1.2.3.4"),
			).UnsafeIgnore(),
			tc("ignore label=apex",
				ignore("@", "", ""),
			).ExpectNoChanges().UnsafeIgnore(),
			tc("ignore type=txt",
				a("@", "1.2.3.4"),
				ignore("", "TXT", ""),
			).ExpectNoChanges().UnsafeIgnore(),
			tc("ignore target=1.2.3.4",
				txt("@", "simple"),
				ignore("", "", "1.2.3.4"),
			).ExpectNoChanges().UnsafeIgnore(),
			tc("ignore manytypes",
				ignore("", "A,TXT", ""),
			).ExpectNoChanges().UnsafeIgnore(),
		).Diff2Only(),

		// Legacy IGNORE_NAME and IGNORE_TARGET tests.

		testgroup("IGNORE_NAME function",
			tc("Create some records",
				txt("foo", "simple"),
				a("foo", "1.2.3.4"),
				a("bar", "1.2.3.4"),
			),
			tc("ignore foo",
				ignoreName("foo"),
				a("bar", "1.2.3.4"),
			).ExpectNoChanges(),
			clear(),
			tc("Create some records",
				txt("bar.foo", "simple"),
				a("bar.foo", "1.2.3.4"),
				a("bar", "1.2.3.4"),
			),
			tc("ignore *.foo",
				ignoreName("*.foo"),
				a("bar", "1.2.3.4"),
			).ExpectNoChanges(),
			clear(),
			tc("Create some records",
				txt("bar.foo", "simple"),
				a("bar.foo", "1.2.3.4"),
			),
			tc("ignore *.foo while we add 1",
				ignoreName("*.foo"),
				a("bar", "1.2.3.4"),
			),
		).Diff2Only(),

		testgroup("IGNORE_NAME apex",
			tc("Create some records",
				txt("@", "simple"),
				a("@", "1.2.3.4"),
				txt("bar", "stringbar"),
				a("bar", "2.4.6.8"),
			).UnsafeIgnore(),
			tc("ignore apex",
				ignoreName("@"),
				txt("bar", "stringbar"),
				a("bar", "2.4.6.8"),
			).ExpectNoChanges().UnsafeIgnore(),
			clear(),
			tc("Add a new record - ignoring apex",
				ignoreName("@"),
				txt("bar", "stringbar"),
				a("bar", "2.4.6.8"),
				a("added", "4.6.8.9"),
			).UnsafeIgnore(),
		).Diff2Only(),

		testgroup("IGNORE_TARGET function CNAME",
			tc("Create some records",
				cname("foo", "test.foo.com."),
				cname("keep", "keep.example.com."),
			),
			tc("ignoring CNAME=test.foo.com.",
				ignoreTarget("test.foo.com.", "CNAME"),
				cname("keep", "keep.example.com."),
			).ExpectNoChanges(),
			tc("ignoring CNAME=test.foo.com. and add",
				ignoreTarget("test.foo.com.", "CNAME"),
				cname("keep", "keep.example.com."),
				a("adding", "1.2.3.4"),
				cname("another", "www.example.com."),
			),
		),

		testgroup("IGNORE_TARGET function CNAME*",
			tc("Create some records",
				cname("foo1", "test.foo.com."),
				cname("foo2", "my.test.foo.com."),
				cname("bar", "test.example.com."),
			).UnsafeIgnore(),
			tc("ignoring CNAME=test.foo.com.",
				ignoreTarget("*.foo.com.", "CNAME"),
				cname("foo2", "my.test.foo.com."),
				cname("bar", "test.example.com."),
			).ExpectNoChanges().UnsafeIgnore(),
			tc("ignoring CNAME=test.foo.com. and add",
				ignoreTarget("*.foo.com.", "CNAME"),
				cname("foo2", "my.test.foo.com."),
				cname("bar", "test.example.com."),
				a("adding", "1.2.3.4"),
				cname("another", "www.example.com."),
			).UnsafeIgnore(),
		),

		testgroup("IGNORE_TARGET function CNAME**",
			tc("Create some records",
				cname("foo1", "test.foo.com."),
				cname("foo2", "my.test.foo.com."),
				cname("bar", "test.example.com."),
			).UnsafeIgnore(),
			tc("ignoring CNAME=test.foo.com.",
				ignoreTarget("**.foo.com.", "CNAME"),
				cname("bar", "test.example.com."),
			).ExpectNoChanges().UnsafeIgnore(),
			tc("ignoring CNAME=test.foo.com. and add",
				ignoreTarget("**.foo.com.", "CNAME"),
				cname("bar", "test.example.com."),
				a("adding", "1.2.3.4"),
				cname("another", "www.example.com."),
			).UnsafeIgnore(),
		),

		// https://github.com/StackExchange/dnscontrol/issues/2285
		// IGNORE_TARGET for CNAMEs wasn't working for AZURE_DNS.
		// Interestingly enough, this has never worked with
		// GANDI_V5/diff1.  It works on all providers in diff2.
		testgroup("IGNORE_TARGET b2285",
			tc("Create some records",
				cname("foo", "redact1.acm-validations.aws."),
				cname("bar", "redact2.acm-validations.aws."),
			),
			tc("Add a new record - ignoring test.foo.com.",
				ignoreTarget("**.acm-validations.aws.", "CNAME"),
			).ExpectNoChanges(),
		).Diff2Only(),

		// Narrative: Congrats! You're done!  If you've made it this far
		// you're very close to being able to submit your PR.  Here's
		// some tips:

		// 1. Ask for help!  It is normal to submit a PR when most (but
		//    not all) tests are passing.  The community would be glad to
		//    help fix the remaining tests.
		// 2. Take a moment to clean up your code. Delete debugging
		//    statements, add comments, run "staticcheck".
		// 3. Thing change: Once your PR is accepted, re-run these tests
		//    every quarter. There may be library updates, API changes,
		//    etc.

	}

	return tests
}
