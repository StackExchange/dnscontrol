package main

// Functions for integration_test.go

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/nameservers"
	"github.com/StackExchange/dnscontrol/v4/pkg/zonerecs"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"github.com/StackExchange/dnscontrol/v4/providers/cloudflare/rtypes/cfsingleredirect"
	"github.com/miekg/dns/dnsutil"
)

var (
	startIdx     = flag.Int("start", -1, "Test number to begin with")
	endIdx       = flag.Int("end", -1, "Test index to stop after")
	verbose      = flag.Bool("verbose", false, "Print corrections as you run them")
	printElapsed = flag.Bool("elapsed", false, "Print elapsed time for each testgroup")
)

// Helper constants/funcs for the CLOUDFLARE proxy testing:

func CfProxyOff() *TestCase   { return tc("proxyoff", cfProxyA("prxy", "174.136.107.111", "off")) }
func CfProxyOn() *TestCase    { return tc("proxyon", cfProxyA("prxy", "174.136.107.111", "on")) }
func CfProxyFull1() *TestCase { return tc("proxyf1", cfProxyA("prxy", "174.136.107.111", "full")) }
func CfProxyFull2() *TestCase { return tc("proxyf2", cfProxyA("prxy", "174.136.107.222", "full")) }
func CfCProxyOff() *TestCase  { return tc("cproxyoff", cfProxyCNAME("cproxy", "example.com.", "off")) }
func CfCProxyOn() *TestCase   { return tc("cproxyon", cfProxyCNAME("cproxy", "example.com.", "on")) }
func CfCProxyFull() *TestCase { return tc("cproxyf", cfProxyCNAME("cproxy", "example.com.", "full")) }

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
func testPermitted(p string, f TestGroup) error {
	// not() and only() can't be mixed.
	if len(f.only) != 0 && len(f.not) != 0 {
		return errors.New("invalid filter: can't mix not() and only()")
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
			if !providers.ProviderHasCapability(*providerFlag, c) {
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
		return errors.New("disabled by only")
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
				_ = rc.SetTarget(strings.Replace(rc.GetTargetField(), "**current-domain**", domainName, 1))
			}
			if strings.Contains(rc.GetLabelFQDN(), "**current-domain**") {
				rc.SetLabelFromFQDN(strings.Replace(rc.GetLabelFQDN(), "**current-domain**", domainName, 1), domainName)
			}

			if strings.Contains(rc.GetTargetField(), "**subscription-id**") {
				_ = rc.SetTarget(strings.Replace(rc.GetTargetField(), "**subscription-id**", origConfig["SubscriptionID"], 1))
			}
			if strings.Contains(rc.GetTargetField(), "**resource-group**") {
				_ = rc.SetTarget(strings.Replace(rc.GetTargetField(), "**resource-group**", origConfig["ResourceGroup"], 1))
			}

			dom.Records = append(dom.Records, &rc)
		}
		dom.Unmanaged = tst.Unmanaged
		dom.UnmanagedUnsafe = tst.UnmanagedUnsafe
		// Bind will refuse a DDNS update when the resulting zone
		// contains a NS record without an associated address
		// records (A or AAAA). In order to run the integration tests
		// against bind, the initial zone contains the following records:
		// - `@ NS dummy-ns.example.com`
		// - `dummy-ns A 9.8.7.6`
		// We 'hardcode' an ignore rule for the `A` record.
		dom.Unmanaged = append(dom.Unmanaged, &models.UnmanagedConfig{
			LabelPattern:  "dummy-ns",
			RTypePattern:  "A",
			TargetPattern: "",
		})
		models.PostProcessRecords(dom.Records)
		dom2, _ := dom.Copy()

		if err := providers.AuditRecords(*providerFlag, dom.Records); err != nil {
			t.Skipf("***SKIPPED(PROVIDER DOES NOT SUPPORT '%s' ::%q)", err, desc)
			return
		}

		// get and run corrections for first time
		_, corrections, actualChangeCount, err := zonerecs.CorrectZoneRecords(prv, dom)
		if err != nil {
			t.Fatal(fmt.Errorf("runTests: %w", err))
		}
		if tst.Changeless {
			if actualChangeCount != 0 {
				t.Logf("Expected 0 corrections on FIRST run, but found %d.", actualChangeCount)
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
		_, corrections, actualChangeCount, err = zonerecs.CorrectZoneRecords(prv, dom2)
		if err != nil {
			t.Fatal(err)
		}
		if actualChangeCount != 0 {
			t.Logf("Expected 0 corrections on second run, but found %d.", actualChangeCount)
			for i, c := range corrections {
				t.Logf("UNEXPECTED #%d: %s", i, c.Msg)
			}
			t.FailNow()
		}
	})
}

func runTests(t *testing.T, prv providers.DNSServiceProvider, domainName string, origConfig map[string]string) {
	dc := getDomainConfigWithNameservers(t, prv, domainName)
	testGroups := makeTests()

	firstGroup := *startIdx
	if firstGroup == -1 {
		firstGroup = 0
	}
	lastGroup := *endIdx
	if lastGroup == -1 {
		lastGroup = len(testGroups)
	}

	curGroup := -1
	for gIdx, group := range testGroups {
		// Abide by -start -end flags
		curGroup++
		if curGroup < firstGroup || curGroup > lastGroup {
			continue
		}

		// Abide by filter
		// fmt.Printf("DEBUG testPermitted: prov=%q profile=%q\n", *providerFlag, *profileFlag)
		if err := testPermitted(*profileFlag, *group); err != nil {
			// t.Logf("%s: ***SKIPPED(%v)***", group.Desc, err)
			makeChanges(t, prv, dc, tc("Empty"), fmt.Sprintf("%02d:%s ***SKIPPED(%v)***", gIdx, group.Desc, err), false, origConfig)
			continue
		}

		// Start the testgroup with a clean slate.
		makeChanges(t, prv, dc, tc("Empty"), "Clean Slate", false, nil)

		// Run the tests.
		start := time.Now()

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

		elapsed := time.Since(start)
		if *printElapsed {
			fmt.Printf("ELAPSED %02d %7.2f %q\n", gIdx, elapsed.Seconds(), group.Desc)
		}
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
	Desc            string
	Records         []*models.RecordConfig
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

func SetLabel(r *models.RecordConfig, label, domain string) {
	r.Name = label
	r.NameFQDN = dnsutil.AddOrigin(label, "**current-domain**.")
}

func withMeta(record *models.RecordConfig, metadata map[string]string) *models.RecordConfig {
	record.Metadata = metadata
	return record
}

func a(name, target string) *models.RecordConfig {
	return makeRec(name, target, "A")
}

func aaaa(name, target string) *models.RecordConfig {
	return makeRec(name, target, "AAAA")
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

func caa(name string, flag uint8, tag string, target string) *models.RecordConfig {
	r := makeRec(name, target, "CAA")
	panicOnErr(r.SetTargetCAA(flag, tag, target))
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

func cfSingleRedirectEnabled() bool {
	return ((*enableCFRedirectMode) != "")
}

func cfSingleRedirect(name string, code any, when, then string) *models.RecordConfig {
	r := makeRec("@", name, cfsingleredirect.SINGLEREDIRECT)
	panicOnErr(cfsingleredirect.FromRaw(r, []any{name, code, when, then})) // Should not happen
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

func aghAPassthrough(pattern, target string) *models.RecordConfig {
	r := makeRec(pattern, target, "ADGUARDHOME_A_PASSTHROUGH")
	return r
}

func aghAAAAPassthrough(pattern, target string) *models.RecordConfig {
	r := makeRec(pattern, target, "ADGUARDHOME_AAAA_PASSTHROUGH")
	return r
}

func cname(name, target string) *models.RecordConfig {
	return makeRec(name, target, "CNAME")
}

func dhcid(name, target string) *models.RecordConfig {
	return makeRec(name, target, "DHCID")
}

func dname(name, target string) *models.RecordConfig {
	return makeRec(name, target, "DNAME")
}

func ds(name string, keyTag uint16, algorithm, digestType uint8, digest string) *models.RecordConfig {
	r := makeRec(name, "", "DS")
	panicOnErr(r.SetTargetDS(keyTag, algorithm, digestType, digest))
	return r
}

func dnskey(name string, flags uint16, protocol, algorithm uint8, publicKey string) *models.RecordConfig {
	r := makeRec(name, "", "DNSKEY")
	panicOnErr(r.SetTargetDNSKEY(flags, protocol, algorithm, publicKey))
	return r
}

func https(name string, priority uint16, target string, params string) *models.RecordConfig {
	r := makeRec(name, target, "HTTPS")
	r.SvcPriority = priority
	r.SvcParams = params
	return r
}

func ignoreName(labelSpec string) *models.RecordConfig {
	return ignore(labelSpec, "*", "*")
}

func ignoreTarget(targetSpec string, typeSpec string) *models.RecordConfig {
	return ignore("*", typeSpec, targetSpec)
}

func ignore(labelSpec string, typeSpec string, targetSpec string) *models.RecordConfig {
	r := &models.RecordConfig{
		Type:     "IGNORE",
		Metadata: map[string]string{},
	}

	r.Metadata["ignore_LabelPattern"] = labelSpec
	r.Metadata["ignore_RTypePattern"] = typeSpec
	r.Metadata["ignore_TargetPattern"] = targetSpec
	return r
}

func loc(name string, d1 uint8, m1 uint8, s1 float32, ns string,
	d2 uint8, m2 uint8, s2 float32, ew string, al float32, sz float32, hp float32, vp float32,
) *models.RecordConfig {
	r := makeRec(name, "", "LOC")
	panicOnErr(r.SetLOCParams(d1, m1, s1, ns, d2, m2, s2, ew, al, sz, hp, vp))
	return r
}

func makeRec(name, target, typ string) *models.RecordConfig {
	r := &models.RecordConfig{
		Type: typ,
		TTL:  300,
	}
	SetLabel(r, name, "**current-domain**.")
	r.MustSetTarget(target)
	return r
}

func manyA(namePattern, target string, n int) []*models.RecordConfig {
	recs := []*models.RecordConfig{}
	for i := range n {
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
	panicOnErr(r.SetTargetNAPTR(order, preference, flags, service, regexp, target))
	return r
}

func ptr(name, target string) *models.RecordConfig {
	return makeRec(name, target, "PTR")
}

func r53alias(name, aliasType, target, evalTargetHealth string) *models.RecordConfig {
	r := makeRec(name, target, "R53_ALIAS")
	r.R53Alias = map[string]string{
		"type":                   aliasType,
		"evaluate_target_health": evalTargetHealth,
	}
	return r
}

func soa(name string, ns, mbox string, serial, refresh, retry, expire, minttl uint32) *models.RecordConfig {
	r := makeRec(name, "", "SOA")
	panicOnErr(r.SetTargetSOA(ns, mbox, serial, refresh, retry, expire, minttl))
	return r
}

func srv(name string, priority, weight, port uint16, target string) *models.RecordConfig {
	r := makeRec(name, target, "SRV")
	panicOnErr(r.SetTargetSRV(priority, weight, port, target))
	return r
}

func sshfp(name string, algorithm uint8, fingerprint uint8, target string) *models.RecordConfig {
	r := makeRec(name, target, "SSHFP")
	panicOnErr(r.SetTargetSSHFP(algorithm, fingerprint, target))
	return r
}

func svcb(name string, priority uint16, target string, params string) *models.RecordConfig {
	r := makeRec(name, target, "SVCB")
	r.SvcPriority = priority
	r.SvcParams = params
	return r
}

func ovhdkim(name, target string) *models.RecordConfig {
	return makeOvhNativeRecord(name, target, "DKIM")
}

func ovhspf(name, target string) *models.RecordConfig {
	return makeOvhNativeRecord(name, target, "SPF")
}

func ovhdmarc(name, target string) *models.RecordConfig {
	return makeOvhNativeRecord(name, target, "DMARC")
}

func makeOvhNativeRecord(name, target, rType string) *models.RecordConfig {
	r := makeRec(name, "", "TXT")
	r.Metadata = make(map[string]string)
	r.Metadata["create_ovh_native_record"] = rType
	r.MustSetTarget(target)
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
	var unmanagedItems []*models.UnmanagedConfig
	for _, r := range recs {
		if r == nil {
			continue
		}
		switch r.Type {
		case "IGNORE":
			unmanagedItems = append(unmanagedItems, &models.UnmanagedConfig{
				LabelPattern:  r.Metadata["ignore_LabelPattern"],
				RTypePattern:  r.Metadata["ignore_RTypePattern"],
				TargetPattern: r.Metadata["ignore_TargetPattern"],
			})
			continue
		default:
			records = append(records, r)
		}
	}
	return &TestCase{
		Desc:      desc,
		Records:   records,
		Unmanaged: unmanagedItems,
	}
}

func txt(name, target string) *models.RecordConfig {
	r := makeRec(name, "", "TXT")
	panicOnErr(r.SetTargetTXT(target))
	return r
}

// func (r *models.RecordConfig) ttl(t uint32) *models.RecordConfig {
func ttl(r *models.RecordConfig, t uint32) *models.RecordConfig {
	r.TTL = t
	return r
}

func tlsa(name string, usage, selector, matchingtype uint8, target string) *models.RecordConfig {
	r := makeRec(name, target, "TLSA")
	panicOnErr(r.SetTargetTLSA(usage, selector, matchingtype, target))
	return r
}

func porkbunUrlfwd(name, target, t, includePath, wildcard string) *models.RecordConfig {
	r := makeRec(name, target, "PORKBUN_URLFWD")
	r.Metadata = make(map[string]string)
	r.Metadata["type"] = t
	r.Metadata["includePath"] = includePath
	r.Metadata["wildcard"] = wildcard
	return r
}

func url(name, target string) *models.RecordConfig {
	return makeRec(name, target, "URL")
}

func url301(name, target string) *models.RecordConfig {
	return makeRec(name, target, "URL301")
}

func frame(name, target string) *models.RecordConfig {
	return makeRec(name, target, "FRAME")
}

func tcEmptyZone() *TestCase {
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
