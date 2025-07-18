package main

// Data-driven tests that exercize the DNS Provider APIs.

import (
	"strings"
	"testing"

	"github.com/StackExchange/dnscontrol/v4/providers"
	_ "github.com/StackExchange/dnscontrol/v4/providers/_all"
)

func TestDNSProviders(t *testing.T) {
	provider, domain, cfg := getProvider(t)
	if provider == nil {
		return
	}
	if domain == "" {
		t.Fatal("NO DOMAIN SET!  Exiting!")
	}

	t.Run(domain, func(t *testing.T) {
		runTests(t, provider, domain, cfg)
	})
}

func makeTests() []*TestGroup {
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
	//     not("ROUTE53"), only("GCLOUD"),  // ERROR!
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

	// emptyzone() is the same as tc("Empty").  It removes all records.
	// Each testgroup() begins with tcEmptyZone() automagically. You do not
	// have to include the tcEmptyZone() in each testgroup().

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

		testgroup("AAAA",
			tc("Create AAAA", aaaa("testaaaa", "2607:f8b0:4006:820::2006")),
			tc("Change AAAA target", aaaa("testaaaa", "2607:f8b0:4006:820::2013")),
		),

		// CNAME

		testgroup("CNAME",
			tc("Create a CNAME", cname("testcname", "www.google.com.")),
			tc("Change CNAME target", cname("testcname", "www.yahoo.com.")),
		),

		testgroup("CNAME-short",
			tc("Create a CNAME",
				a("foo", "1.2.3.4"),
				cname("testcname", "foo"),
			),
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
			tc("Create MX apex", mx("@", 5, "foo.com.")),
			tc("Change MX apex", mx("@", 5, "bar.com.")),
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
			tcEmptyZone(),
			tc("Create an A record", a("www", "1.1.1.1")),
			tc("Add at label1", a("www", "1.1.1.1"), a("www", "2.2.2.2")),
			tc("Add at label2", a("www", "1.1.1.1"), a("www", "2.2.2.2"), a("www", "3.3.3.3")),
		),

		testgroup("manyTypesAtOnce",
			tc("CreateManyTypesAtLabel", a("www", "1.1.1.1"), mx("testmx", 5, "foo.com."), mx("testmx", 100, "bar.com.")),
			tcEmptyZone(),
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

		testgroup("HTTPS",
			requires(providers.CanUseHTTPS),
			tc("Create a HTTPS record", https("@", 1, "test.com.", "port=80")),
			tc("Change HTTPS priority", https("@", 2, "test.com.", "port=80")),
			tc("Change HTTPS target", https("@", 2, ".", "port=80")),
			tc("Change HTTPS params", https("@", 2, ".", "port=99")),
			tc("Change HTTPS params-empty", https("@", 2, ".", "")),
			tc("Change HTTPS all", https("@", 3, "example.com.", "port=100")),
		),

		testgroup("Ech",
			requires(providers.CanUseHTTPS),
			tc("Create a HTTPS record", https("@", 1, "example.com.", "alpn=h2,h3")),
			tc("Add an ECH key", https("@", 1, "example.com.", "alpn=h2,h3 ech=some+base64+encoded+value///")),
			tc("Ignore the ECH key while changing other values", https("@", 1, "example.net.", "port=80 ech=IGNORE")),
			// tc("Should be a no-op", https("@", 1, "example.net.", "port=80 ech=some+base64+encoded+value///")),
			tc("Change the ECH key and other values", https("@", 1, "example.org.", "port=80 ipv4hint=127.0.0.1 ech=another+base64+encoded+value")),
			// tc("Ignore the ECH key while not changing anything", https("@", 1, "example.org.", "port=80 ipv4hint=127.0.0.1 ech=IGNORE")),
			// tc("Should be a no-op", https("@", 1, "example.org.", "port=80 ipv4hint=127.0.0.1 ech=another+base64+encoded+value")),
			tc("Another domain with a different ECH value", https("ech", 1, "example.com.", "ech=some+base64+encoded+value///")),
		),

		testgroup("SVCB",
			requires(providers.CanUseSVCB),
			tc("Create a SVCB record", svcb("@", 1, "test.com.", "port=80")),
			tc("Change SVCB priority", svcb("@", 2, "test.com.", "port=80")),
			tc("Change SVCB target", svcb("@", 2, ".", "port=80")),
			tc("Change SVCB params", svcb("@", 2, ".", "port=99")),
			tc("Change SVCB params-empty", svcb("@", 2, ".", "")),
			tc("Change SVCB all", svcb("@", 3, "example.com.", "port=100")),
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
			tc("Record pointing to @",
				cname("foo", "**current-domain**."),
				a("@", "1.2.3.4"),
			),
		),

		testgroup("ApexMX",
			tc("Record pointing to @",
				mx("foo", 8, "**current-domain**."),
				a("@", "1.2.3.4"),
			),
		),

		// RFC 7505 NullMX
		testgroup("NullMX",
			not(
				"TRANSIP", // TRANSIP is slow and doesn't support NullMX. Skip to save time.
			),
			tc("create", // Install a Null MX.
				a("nmx", "1.2.3.3"), // Install this so it is ready for the next tc()
				a("www", "1.2.3.9"), // Install this so it is ready for the next tc()
				mx("nmx", 0, "."),
			),
			tc("unnull", // Change to regular MX.
				a("nmx", "1.2.3.3"),
				a("www", "1.2.3.9"),
				mx("nmx", 3, "nmx.**current-domain**."),
				mx("nmx", 9, "www.**current-domain**."),
			),
			tc("renull", // Change back to Null MX.
				a("nmx", "1.2.3.3"),
				a("www", "1.2.3.9"),
				mx("nmx", 0, "."),
			),
		),

		// RFC 7505 NullMX at Apex
		testgroup("NullMXApex",
			not(
				"TRANSIP", // TRANSIP is slow and doesn't support NullMX. Skip to save time.
			),
			tc("create", // Install a Null MX.
				a("@", "1.2.3.2"),   // Install this so it is ready for the next tc()
				a("www", "1.2.3.8"), // Install this so it is ready for the next tc()
				mx("@", 0, "."),
			),
			tc("unnull", // Change to regular MX.
				a("@", "1.2.3.2"),
				a("www", "1.2.3.8"),
				mx("@", 2, "**current-domain**."),
				mx("@", 8, "www.**current-domain**."),
			),
			tc("renull", // Change back to Null MX.
				a("@", "1.2.3.2"),
				a("www", "1.2.3.8"),
				mx("@", 0, "."),
			),
		),

		testgroup("NS",
			not(
				"DNSIMPLE",  // Does not support NS records nor subdomains.
				"EXOSCALE",  // Not supported.
				"NETCUP",    // NS records not currently supported.
				"FORTIGATE", // Not supported
			),
			tc("NS for subdomain", ns("xyz", "ns2.foo.com.")),
			tc("Dual NS for subdomain", ns("xyz", "ns2.foo.com."), ns("xyz", "ns1.foo.com.")),
			tc("NS Record pointing to @", a("@", "1.2.3.4"), ns("foo", "**current-domain**.")),
		),

		testgroup("NS only APEX",
			not(
				"DNSIMPLE", // Does not support NS records nor subdomains.
				"EXOSCALE", // Not supported.
				"NETCUP",   // NS records not currently supported.
			),
			tc("Single NS at apex", ns("@", "ns1.foo.com.")),
			tc("Dual NS at apex", ns("@", "ns2.foo.com."), ns("@", "ns1.foo.com.")),
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
			// of record.

			// Some of these test cases are commented out because they test
			// something that isn't widely used or supported.  For example
			// many APIs don't support a backslack (`\`) in a TXT record;
			// luckily we've never seen a need for that "in the wild".  If
			// you want to future-proof your provider, temporarily remove
			// the comments and get those tests working, or reject it using
			// auditrecords.go.

			// ProTip: Unsure how a provider's API escapes something? Try
			// adding the TXT record via the Web UI and watch how the string
			// is escaped when you download the records.

			// Nobody needs this and many APIs don't allow it.
			tc("a 0-byte TXT", txt("foo0", "")),

			// Test edge cases around 255, 255*2, 255*3:
			tc("a 254-byte TXT", txt("foo254", strings.Repeat("A", 254))), // 255-1
			tc("a 255-byte TXT", txt("foo255", strings.Repeat("B", 255))), // 255
			tc("a 256-byte TXT", txt("foo256", strings.Repeat("C", 256))), // 255+1
			tc("a 509-byte TXT", txt("foo509", strings.Repeat("D", 509))), // 255*2-1
			tc("a 510-byte TXT", txt("foo510", strings.Repeat("E", 510))), // 255*2
			tc("a 511-byte TXT", txt("foo511", strings.Repeat("F", 511))), // 255*2+1
			tc("a 764-byte TXT", txt("foo764", strings.Repeat("G", 764))), // 255*3-1
			tc("a 765-byte TXT", txt("foo765", strings.Repeat("H", 765))), // 255*3
			tc("a 766-byte TXT", txt("foo766", strings.Repeat("J", 766))), // 255*3+1
			// tcEmptyZone(),

			tc("TXT with 1 single-quote", txt("foosq", "quo'te")),
			tc("TXT with 1 backtick", txt("foobt", "blah`blah")),
			tc("TXT with 1 dq-1interior", txt("foodq", `in"side`)),
			tc("TXT with 2 dq-2interior", txt("foodqs", `in"ter"ior`)),
			tc("TXT with 1 dq-left", txt("foodqs", `"left`)),
			tc("TXT with 1 dq-right", txt("foodqs", `right"`)),

			// Semicolons don't need special treatment.
			// https://serverfault.com/questions/743789
			tc("TXT with semicolon", txt("foosc1", `semi;colon`)),
			tc("TXT with semicolon ws", txt("foosc2", `wssemi ; colon`)),

			tc("TXT interior ws", txt("foosp", "with spaces")),
			// tc("TXT leading ws", txt("foowsb", " leadingspace")),
			tc("TXT trailing ws", txt("foows1", "trailingws ")),

			// Vultr syntax-checks TXT records with SPF contents.
			tc("Create a TXT/SPF", txt("foo", "v=spf1 ip4:99.99.99.99 -all")),

			// Nobody needs this and many APIs don't allow it.
			// tc("Create TXT with frequently difficult characters", txt("fooex", `!^.*$@#%^&()([][{}{<></:;-_=+\`)),
		),

		testgroup("TXT backslashes",
			tc("TXT with backslashs",
				txt("fooosbs1", `1back\slash`),
				txt("fooosbs2", `2back\\slash`),
				txt("fooosbs3", `3back\\\slash`),
				txt("fooosbs4", `4back\\\\slash`)),
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
				// a("foo", "3.4.5.6"), // Delete
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
				mx("foo", 10, "foo.**current-domain**."),
				mx("foo", 20, "bar.**current-domain**."),
			),
			tc("changeOne",
				a("bar", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				a("foo", "8.8.8.8"), // Change
				mx("foo", 10, "foo.**current-domain**."),
				mx("foo", 20, "bar.**current-domain**."),
			),
			tc("deleteOne",
				a("bar", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				// a("foo", "8.8.8.8"),  // Delete
				mx("foo", 10, "foo.**current-domain**."),
				mx("foo", 20, "bar.**current-domain**."),
			),
			tc("addOne",
				a("bar", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				a("foo", "8.8.8.8"), // Add
				mx("foo", 10, "foo.**current-domain**."),
				mx("foo", 20, "bar.**current-domain**."),
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
				"DESEC",         // Skip due to daily update limits.
				// "CSCGLOBAL",     // Doesn't page. Works fine.  Due to the slow API we skip.
				"GANDI_V5",   // Their API is so damn slow. We'll add it back as needed.
				"HEDNS",      // Doesn't page. Works fine.  Due to the slow API we skip.
				"HEXONET",    // Doesn't page. Works fine.  Due to the slow API we skip.
				"LOOPIA",     // Their API is so damn slow. Plus, no paging.
				"NAMEDOTCOM", // Their API is so damn slow. We'll add it back as needed.
				"NS1",        // Free acct only allows 50 records, therefore we skip
				// "ROUTE53",       // Batches up changes in pages.
				"TRANSIP",   // Doesn't page. Works fine.  Due to the slow API we skip.
				"CNR",       // Test beaks limits.
				"FORTIGATE", // No paging
			),
			tc("99 records", manyA("pager101-rec%04d", "1.2.3.4", 99)...),
			tc("100 records", manyA("pager101-rec%04d", "1.2.3.4", 100)...),
			tc("101 records", manyA("pager101-rec%04d", "1.2.3.4", 101)...),
		),

		testgroup("pager601",
			only(
				// "AZURE_DNS",     // Removed because it is too slow
				//"CLOUDFLAREAPI", // Infinite pagesize but due to slow speed, skipping.
				//"CSCGLOBAL",     // Doesn't page. Works fine.  Due to the slow API we skip.
				//"DESEC",         // Skip due to daily update limits.
				//"GANDI_V5",      // Their API is so damn slow. We'll add it back as needed.
				//"GCLOUD",
				//"HEXONET", // Doesn't page. Works fine.  Due to the slow API we skip.
				"ROUTE53", // Batches up changes in pages.
			),
			tc("601 records", manyA("pager601-rec%04d", "1.2.3.4", 600)...),
			tc("Update 601 records", manyA("pager601-rec%04d", "1.2.3.5", 600)...),
		),

		testgroup("pager1201",
			only(
				// "AKAMAIEDGEDNS", // No paging done. No need to test.
				//"AZURE_DNS",     // Currently failing. See https://github.com/StackExchange/dnscontrol/issues/770
				//"CLOUDFLAREAPI", // Fails with >1000 corrections. See https://github.com/StackExchange/dnscontrol/issues/1440
				//"CSCGLOBAL",     // Doesn't page. Works fine.  Due to the slow API we skip.
				//"DESEC",         // Skip due to daily update limits.
				//"GANDI_V5",      // Their API is so damn slow. We'll add it back as needed.
				//"HEDNS",         // No paging done. No need to test.
				//"GCLOUD",
				//"HEXONET", // Doesn't page. Works fine.  Due to the slow API we skip.
				"HOSTINGDE", // Pages.
				"ROUTE53",   // Batches up changes in pages.
			),
			tc("1200 records", manyA("pager1201-rec%04d", "1.2.3.4", 1200)...),
			tc("Update 1200 records", manyA("pager1201-rec%04d", "1.2.3.5", 1200)...),
		),

		// Test the boundaries of Google' batch system.
		// 1200 is used because it is larger than batchMax.
		// https://github.com/StackExchange/dnscontrol/pull/2762#issuecomment-1877825559
		testgroup("batchRecordswithOthers",
			only(
				//"GCLOUD",
				"HOSTINGDE", // Pages.
			),
			tc("1200 records",
				manyA("batch-rec%04d", "1.2.3.4", 1200)...),
			tc("Update 1200 records and Create others", append(
				manyA("batch-arec%04d", "1.2.3.4", 1200),
				manyA("batch-rec%04d", "1.2.3.5", 1200)...)...),
			tc("Update 1200 records and Create and Delete others", append(
				manyA("batch-rec%04d", "1.2.3.4", 1200),
				manyA("batch-zrec%04d", "1.2.3.4", 1200)...)...),
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
			tc("CAA record", caa("@", 0, "issue", "letsencrypt.org")),
			tc("CAA change tag", caa("@", 0, "issuewild", "letsencrypt.org")),
			tc("CAA change target", caa("@", 0, "issuewild", "example.com")),
			tc("CAA change flag", caa("@", 128, "issuewild", "example.com")),
			tc("CAA many records", caa("@", 128, "issuewild", ";")),
			// Test support of spaces in the 3rd field. Some providers don't
			// support this.  See providers/exoscale/auditrecords.go as an example.
			tc("CAA whitespace", caa("@", 0, "issue", "letsencrypt.org; validationmethods=dns-01; accounturi=https://acme-v02.api.letsencrypt.org/acme/acct/1234")),
		),

		// LOCation records. // No.47
		testgroup("LOC",
			requires(providers.CanUseLOC),
			// 42 21 54     N  71 06  18     W -24m 30m
			tc("Single LOC record", loc("@", 42, 21, 54, "N", 71, 6, 18, "W", -24.05, 30, 0, 0)),
			// 42 21 54     N  71 06  18     W -24m 30m
			tc("Update single LOC record", loc("@", 42, 21, 54, "N", 71, 6, 18, "W", -24.06, 30, 10, 0)),
			tc("Multiple LOC records-create a-d modify apex", // create a-d, modify @
				// 42 21 54     N  71 06  18     W -24m 30m
				loc("@", 42, 21, 54, "N", 71, 6, 18, "W", -24, 30, 0, 0),
				// 42 21 43.952 N  71 5   6.344  W -24m 1m 200m
				loc("a", 42, 21, 43.952, "N", 71, 5, 6.344, "W", -24.33, 1, 200, 10),
				// 52 14 05     N  00 08  50     E 10m
				loc("b", 52, 14, 5, "N", 0, 8, 50, "E", 10.22, 0, 0, 0),
				// 32  7 19     S 116  2  25     E 10m
				loc("c", 32, 7, 19, "S", 116, 2, 25, "E", 10, 0, 0, 0),
				// 42 21 28.764 N  71 00  51.617 W -44m 2000m
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
			not("CLOUDNS",
				"FORTIGATE", // FortiGate does not really support ARPA Zones and handles PTR records really weired
			),
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
			tcEmptyZone(), // Required or only the first run passes.
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
			tcEmptyZone(),
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
			not("CLOUDFLAREAPI"),
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
			// tc("modify field 2", ds("child", 65535, 254, 4, "0123456789ABCDEF")),
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
			// tc("delete all DS",
			//	ns("another-child", "ns101.cloudns.net."),
			//),
		),
		testgroup("DHCID",
			requires(providers.CanUseDHCID),
			tc("Create DHCID record", dhcid("test", "AAIBY2/AuCccgoJbsaxcQc9TUapptP69lOjxfNuVAA2kjEA=")),
			tc("Modify DHCID record", dhcid("test", "AAAAAAAAuCccgoJbsaxcQc9TUapptP69lOjxfNuVAA2kjEA=")),
		),

		testgroup("DNAME",
			requires(providers.CanUseDNAME),
			tc("Create DNAME record", dname("test", "example.com.")),
			tc("Modify DNAME record", dname("test", "example.net.")),
			tc("Create DNAME record in non-FQDN", dname("a", "b")),
		),

		testgroup("DNSKEY",
			requires(providers.CanUseDNSKEY),
			tc("Create DNSKEY record", dnskey("test", 257, 3, 13, "fRnjbeUVyKvz1bDx2lPmu3KY1k64T358t8kP6Hjveos=")),
			tc("Modify DNSKEY record 1", dnskey("test", 256, 3, 13, "fRnjbeUVyKvz1bDx2lPmu3KY1k64T358t8kP6Hjveos=")),
			tc("Modify DNSKEY record 2", dnskey("test", 256, 3, 13, "whjtMiJP9C86l0oTJUxemuYtQ0RIZePWt6QETC2kkKM=")),
			tc("Modify DNSKEY record 3", dnskey("test", 256, 3, 15, "whjtMiJP9C86l0oTJUxemuYtQ0RIZePWt6QETC2kkKM=")),
		),

		//// Vendor-specific record types

		// Narrative: DNSControl supports DNS records that don't exist!
		// Well, they exist for particular vendors.  Let's test each of
		// them here. If you are writing a new provider, I have some good
		// news: These don't apply to you!

		testgroup("ALIAS on apex",
			requires(providers.CanUseAlias),
			tc("ALIAS at root", alias("@", "foo.com.")),
			tc("change it", alias("@", "foo2.com.")),
		),

		testgroup("ALIAS to nonfqdn",
			requires(providers.CanUseAlias),
			tc("ALIAS at root",
				a("foo", "1.2.3.4"),
				alias("@", "foo"),
			),
		),

		testgroup("ALIAS on subdomain",
			requires(providers.CanUseAlias),
			not("TRANSIP"), // TransIP does support ALIAS records, but only for apex records (@)
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
				azureAlias("bar.a", "A", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain**/A/foo.a"),
			),
			tc("change aliasA",
				a("foo.a", "1.2.3.4"),
				a("quux.a", "2.3.4.5"),
				azureAlias("bar.a", "A", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain**/A/quux.a"),
			),
			tc("change backA",
				a("foo.a", "1.2.3.4"),
				a("quux.a", "2.3.4.5"),
				azureAlias("bar.a", "A", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain**/A/foo.a"),
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
				azureAlias("bar.cname", "CNAME", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain**/CNAME/foo.cname"),
			),
			tc("change aliasCNAME",
				cname("foo.cname", "google.com."),
				cname("quux.cname", "google2.com."),
				azureAlias("bar.cname", "CNAME", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain**/CNAME/quux.cname"),
			),
			tc("change backCNAME",
				cname("foo.cname", "google.com."),
				cname("quux.cname", "google2.com."),
				azureAlias("bar.cname", "CNAME", "/subscriptions/**subscription-id**/resourceGroups/**resource-group**/providers/Microsoft.Network/dnszones/**current-domain**/CNAME/foo.cname"),
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
				r53alias("kenny", "A", "kyle.**current-domain**.", "false"),
			),
			tc("modify an r53 alias",
				a("kyle", "1.2.3.4"),
				a("cartman", "2.3.4.5"),
				r53alias("kenny", "A", "cartman.**current-domain**.", "false"),
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
				r53alias("dev-system", "CNAME", "dev-system18.**current-domain**.", "false"),
			),
			tc("modify alias to 19",
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
				cname("dev-system19", "ec2-54-91-99-999.compute-1.amazonaws.com."),
				r53alias("dev-system", "CNAME", "dev-system19.**current-domain**.", "false"),
			),
			tc("remove alias",
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
				cname("dev-system19", "ec2-54-91-99-999.compute-1.amazonaws.com."),
			),
			tc("add an alias back",
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
				cname("dev-system19", "ec2-54-91-99-999.compute-1.amazonaws.com."),
				r53alias("dev-system", "CNAME", "dev-system19.**current-domain**.", "false"),
			),
			tc("remove cnames",
				r53alias("dev-system", "CNAME", "dev-system19.**current-domain**.", "false"),
			),
		),

		testgroup("R53_ALIAS_CNAME",
			requires(providers.CanUseRoute53Alias),
			tc("create alias+cname in one step",
				r53alias("dev-system", "CNAME", "dev-system18.**current-domain**.", "false"),
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
			),
		),

		testgroup("R53_ALIAS_Loop",
			// This will always be skipped because rejectifTargetEqualsLabel
			// will always flag it as not permitted.
			// See https://github.com/StackExchange/dnscontrol/issues/2107
			requires(providers.CanUseRoute53Alias),
			tc("loop should fail",
				r53alias("test-islandora", "CNAME", "test-islandora.**current-domain**.", "false"),
			),
		),

		// Bug https://github.com/StackExchange/dnscontrol/issues/2285
		testgroup("R53_alias pre-existing",
			requires(providers.CanUseRoute53Alias),
			tc("Create some records",
				r53alias("dev-system", "CNAME", "dev-system18.**current-domain**.", "false"),
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
			),
			tc("Add a new record - ignoring foo",
				a("bar", "1.2.3.4"),
				ignoreName("dev-system*"),
			),
		),

		testgroup("R53_alias evaluate_target_health",
			requires(providers.CanUseRoute53Alias),
			tc("Create alias and cname",
				r53alias("test-record", "CNAME", "test-record-1.**current-domain**.", "false"),
				cname("test-record-1", "ec2-54-91-33-155.compute-1.amazonaws.com."),
			),
			tc("modify evaluate target health",
				r53alias("test-record", "CNAME", "test-record-1.**current-domain**.", "true"),
				cname("test-record-1", "ec2-54-91-33-155.compute-1.amazonaws.com."),
			),
		),

		// Bug https://github.com/StackExchange/dnscontrol/issues/3493
		// Summary: R53_ALIAS -> CNAME conversion doesn't work.
		testgroup("R53_B3493",
			requires(providers.CanUseRoute53Alias),
			// Create the R53_ALIAS:
			tc("b3493 create alias+cname in one step",
				r53alias("dev-system", "CNAME", "dev-system18.**current-domain**.", "false"),
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
			),
			// Convert R53_ALIAS -> CNAME.
			tc("convert r53alias to cname",
				cname("dev-system", "dev-system18.**current-domain**."),
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
			),
		),
		// Verify CNAME -> R53_ALIAS works too. (not part of the bug, but worth verifying)
		testgroup("R53_B3493_REV",
			requires(providers.CanUseRoute53Alias),
			// Create the CNAME
			tc("b3493 create cnames",
				cname("dev-system", "dev-system18.**current-domain**."),
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
			),
			// Convert CNAME -> R53_ALIAS.
			tc("convert cname to r53_alias",
				r53alias("dev-system", "CNAME", "dev-system18.**current-domain**.", "false"),
				cname("dev-system18", "ec2-54-91-33-155.compute-1.amazonaws.com."),
			),
		),

		// CLOUDFLAREAPI features

		// CLOUDFLAREAPI: Redirects:

		// go test -v -verbose -profile CLOUDFLAREAPI                // PAGE_RULEs
		// go test -v -verbose -profile CLOUDFLAREAPI -cfredirect=c  // Convert: Convert page rules to Single Redirect
		// go test -v -verbose -profile CLOUDFLAREAPI -cfredirect=n  // New: Convert old to new Single Redirect
		// ProTip: Add this to just run this test:
		//  -start 59 -end 60

		testgroup("CF_REDIRECT",
			only("CLOUDFLAREAPI"),
			tc("redir", cfRedir("cnn.**current-domain**/*", "https://www.cnn.com/$1")),
			tc("change", cfRedir("cnn.**current-domain**/*", "https://change.cnn.com/$1")),
			tc("changelabel", cfRedir("cable.**current-domain**/*", "https://change.cnn.com/$1")),

			// Removed these for speed.  They tested if order matters,
			// which it doesn't seem to.  Re-add if needed.
			tcEmptyZone(),
			tc("multipleA",
				cfRedir("cnn.**current-domain**/*", "https://www.cnn.com/$1"),
				cfRedir("msnbc.**current-domain**/*", "https://msnbc.cnn.com/$1"),
			),
			tcEmptyZone(),
			tc("multipleB",
				cfRedir("msnbc.**current-domain**/*", "https://msnbc.cnn.com/$1"),
				cfRedir("cnn.**current-domain**/*", "https://www.cnn.com/$1"),
			),
			tc("change1",
				cfRedir("msnbc.**current-domain**/*", "https://msnbc.cnn.com/$1"),
				cfRedir("cnn.**current-domain**/*", "https://change.cnn.com/$1"),
			),
			tc("change1",
				cfRedir("msnbc.**current-domain**/*", "https://msnbc.cnn.com/$1"),
				cfRedir("cablenews.**current-domain**/*", "https://change.cnn.com/$1"),
			),

			// NB(tlim): This test case used to fail but mysteriously started working.
			tcEmptyZone(),
			tc("multiple3",
				cfRedir("msnbc.**current-domain**/*", "https://msnbc.cnn.com/$1"),
				cfRedir("cnn.**current-domain**/*", "https://www.cnn.com/$1"),
				cfRedir("nytimes.**current-domain**/*", "https://www.nytimes.com/$1"),
			),

			// Repeat the above tests using CF_TEMP_REDIR instead
			tcEmptyZone(),
			tc("tempredir", cfRedirTemp("cnn.**current-domain**/*", "https://www.cnn.com/$1")),
			tc("tempchange", cfRedirTemp("cnn.**current-domain**/*", "https://change.cnn.com/$1")),
			tc("tempchangelabel", cfRedirTemp("cable.**current-domain**/*", "https://change.cnn.com/$1")),
			tcEmptyZone(),
			tc("tempmultipleA",
				cfRedirTemp("cnn.**current-domain**/*", "https://www.cnn.com/$1"),
				cfRedirTemp("msnbc.**current-domain**/*", "https://msnbc.cnn.com/$1"),
			),
			tcEmptyZone(),
			tc("tempmultipleB",
				cfRedirTemp("msnbc.**current-domain**/*", "https://msnbc.cnn.com/$1"),
				cfRedirTemp("cnn.**current-domain**/*", "https://www.cnn.com/$1"),
			),
			tc("tempchange1",
				cfRedirTemp("msnbc.**current-domain**/*", "https://msnbc.cnn.com/$1"),
				cfRedirTemp("cnn.**current-domain**/*", "https://change.cnn.com/$1"),
			),
			tc("tempchange1",
				cfRedirTemp("msnbc.**current-domain**/*", "https://msnbc.cnn.com/$1"),
				cfRedirTemp("cablenews.**current-domain**/*", "https://change.cnn.com/$1"),
			),
			// NB(tlim): This test case used to fail but mysteriously started working.
			tc("tempmultiple3",
				cfRedirTemp("msnbc.**current-domain**/*", "https://msnbc.cnn.com/$1"),
				cfRedirTemp("cnn.**current-domain**/*", "https://www.cnn.com/$1"),
				cfRedirTemp("nytimes.**current-domain**/*", "https://www.nytimes.com/$1"),
			),
		),

		testgroup("CF_REDIRECT_CONVERT",
			only("CLOUDFLAREAPI"),
			alltrue(cfSingleRedirectEnabled()),
			tc("start301", cfRedir("cnn.**current-domain**/*", "https://www.cnn.com/$1")),
			tc("convert302", cfRedirTemp("cnn.**current-domain**/*", "https://www.cnn.com/$1")),
			tc("convert301", cfRedir("cnn.**current-domain**/*", "https://www.cnn.com/$1")),
		),

		testgroup("CLOUDFLAREAPI_SINGLE_REDIRECT",
			only("CLOUDFLAREAPI"),
			alltrue(cfSingleRedirectEnabled()),
			tc("start301", cfSingleRedirect(`name1`, `301`, `http.host eq "cnn.slackoverflow.com"`, `concat("https://www.cnn.com", http.request.uri.path)`)),
			tc("changecode", cfSingleRedirect(`name1`, `302`, `http.host eq "cnn.slackoverflow.com"`, `concat("https://www.cnn.com", http.request.uri.path)`)),
			tc("changewhen", cfSingleRedirect(`name1`, `302`, `http.host eq "msnbc.slackoverflow.com"`, `concat("https://www.cnn.com", http.request.uri.path)`)),
			tc("changethen", cfSingleRedirect(`name1`, `302`, `http.host eq "msnbc.slackoverflow.com"`, `concat("https://www.msnbc.com", http.request.uri.path)`)),
			tc("changename", cfSingleRedirect(`name1bis`, `302`, `http.host eq "msnbc.slackoverflow.com"`, `concat("https://www.msnbc.com", http.request.uri.path)`)),
		),

		// CLOUDFLAREAPI: PROXY

		testgroup("CF_PROXY A create",
			only("CLOUDFLAREAPI"),
			CfProxyOff(), tcEmptyZone(),
			CfProxyOn(), tcEmptyZone(),
			CfProxyFull1(), tcEmptyZone(),
			CfProxyFull2(), tcEmptyZone(),
		),

		// These next testgroups attempt every possible transition between off, on, full1 and full2.
		// "full1" simulates "full" without the IP being translated.
		// "full2" simulates "full" WITH the IP translated.

		testgroup("CF_PROXY A off to X",
			only("CLOUDFLAREAPI"),
			// CF_PROXY_OFF(), CF_PROXY_OFF(), tcEmptyZone(), // redundant
			CfProxyOff(), CfProxyOn(), tcEmptyZone(),
			CfProxyOff(), CfProxyFull1(), tcEmptyZone(),
			CfProxyOff(), CfProxyFull2(), tcEmptyZone(),
		),

		testgroup("CF_PROXY A on to X",
			only("CLOUDFLAREAPI"),
			CfProxyOn(), CfProxyOff(), tcEmptyZone(),
			// CF_PROXY_ON(), CF_PROXY_ON(), tcEmptyZone(), // redundant
			// CF_PROXY_ON(), CF_PROXY_FULL1().ExpectNoChanges(), tcEmptyZone(), // Removed for speed
			CfProxyOn(), CfProxyFull2(), tcEmptyZone(),
		),

		testgroup("CF_PROXY A full1 to X",
			only("CLOUDFLAREAPI"),
			CfProxyFull1(), CfProxyOff(), tcEmptyZone(),
			// CF_PROXY_FULL1(), CF_PROXY_ON().ExpectNoChanges(), tcEmptyZone(), // Removed for speed
			// CF_PROXY_FULL1(), tcEmptyZone(), // redundant
			CfProxyFull1(), CfProxyFull2(), tcEmptyZone(),
		),

		testgroup("CF_PROXY A full2 to X",
			only("CLOUDFLAREAPI"),
			CfProxyFull2(), CfProxyOff(), tcEmptyZone(),
			CfProxyFull2(), CfProxyOn(), tcEmptyZone(),
			CfProxyFull2(), CfProxyFull1(), tcEmptyZone(),
			// CF_PROXY_FULL2(), CF_PROXY_FULL2(), tcEmptyZone(), // redundant
		),

		testgroup("CF_PROXY CNAME create",
			only("CLOUDFLAREAPI"),
			CfCProxyOff(), tcEmptyZone(),
			CfCProxyOn(), tcEmptyZone(),
			CfCProxyFull(), tcEmptyZone(),
		),

		testgroup("CF_PROXY CNAME off to X",
			only("CLOUDFLAREAPI"),
			// CF_CPROXY_OFF(), CF_CPROXY_OFF(), tcEmptyZone(),  // redundant
			CfCProxyOff(), CfCProxyOn(), tcEmptyZone(),
			CfCProxyOff(), CfCProxyFull(), tcEmptyZone(),
		),

		testgroup("CF_PROXY CNAME on to X",
			only("CLOUDFLAREAPI"),
			CfCProxyOn(), CfCProxyOff(), tcEmptyZone(),
			// CF_CPROXY_ON(), CF_CPROXY_ON(), tcEmptyZone(), // redundant
			// CF_CPROXY_ON(), CF_CPROXY_FULL().ExpectNoChanges(), tcEmptyZone(), // Removed for speed
		),

		testgroup("CF_PROXY CNAME full to X",
			only("CLOUDFLAREAPI"),
			CfCProxyFull(), CfCProxyOff(), tcEmptyZone(),
			// CF_CPROXY_FULL(), CF_CPROXY_ON().ExpectNoChanges(), tcEmptyZone(), // Removed for speed
			// CF_CPROXY_FULL(), tcEmptyZone(), // redundant
		),

		testgroup("CF_WORKER_ROUTE",
			only("CLOUDFLAREAPI"),
			alltrue(*enableCFWorkers),
			// TODO(fdcastel): Add worker scripts via api call before test execution
			tc("simple", cfWorkerRoute("cnn.**current-domain**/*", "dnscontrol_integrationtest_cnn")),
			tc("changeScript", cfWorkerRoute("cnn.**current-domain**/*", "dnscontrol_integrationtest_msnbc")),
			tc("changePattern", cfWorkerRoute("cable.**current-domain**/*", "dnscontrol_integrationtest_msnbc")),
			tcEmptyZone(),
			tc("createMultiple",
				cfWorkerRoute("cnn.**current-domain**/*", "dnscontrol_integrationtest_cnn"),
				cfWorkerRoute("msnbc.**current-domain**/*", "dnscontrol_integrationtest_msnbc"),
			),
			tc("addOne",
				cfWorkerRoute("msnbc.**current-domain**/*", "dnscontrol_integrationtest_msnbc"),
				cfWorkerRoute("cnn.**current-domain**/*", "dnscontrol_integrationtest_cnn"),
				cfWorkerRoute("api.**current-domain**/cnn/*", "dnscontrol_integrationtest_cnn"),
			),
			tc("changeOne",
				cfWorkerRoute("msn.**current-domain**/*", "dnscontrol_integrationtest_msnbc"),
				cfWorkerRoute("cnn.**current-domain**/*", "dnscontrol_integrationtest_cnn"),
				cfWorkerRoute("api.**current-domain**/cnn/*", "dnscontrol_integrationtest_cnn"),
			),
			tc("deleteOne",
				cfWorkerRoute("msn.**current-domain**/*", "dnscontrol_integrationtest_msnbc"),
				cfWorkerRoute("api.**current-domain**/cnn/*", "dnscontrol_integrationtest_cnn"),
			),
		),

		testgroup("ADGUARDHOME_A_PASSTHROUGH",
			only("ADGUARDHOME"),
			tc("simple", aghAPassthrough("foo", "")),
		),

		testgroup("ADGUARDHOME_AAAA_PASSTHROUGH",
			only("ADGUARDHOME"),
			tc("simple", aghAAAAPassthrough("foo", "")),
		),

		//// IGNORE* features

		// Narrative: You're basically done now. These remaining tests
		// exercise the NO_PURGE and IGNORE* features.  These are handled
		// by the pkg/diff2 module. If they work for any provider, they
		// should work for all providers.  However we're going to test
		// them anyway because one never knows.  Ready?  Let's go!

		testgroup("IGNORE main",
			tc("Create some records",
				a("foo", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			),

			tc("ignore label",
				// NB(tlim): This ignores 1 record of a recordSet. This should
				// fail for diff2.ByRecordSet() providers if diff2 is not
				// implemented correctly.
				// a("foo", "1.2.3.4"),
				// a("foo", "2.3.4.5"),
				// txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
				ignore("foo", "", ""),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("foo", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			tc("ignore label,type",
				// a("foo", "1.2.3.4"),
				// a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
				ignore("foo", "A", ""),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("foo", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			tc("ignore label,type,target",
				// a("foo", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
				ignore("foo", "A", "1.2.3.4"),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("foo", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			tc("ignore type",
				// a("foo", "1.2.3.4"),
				// a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				// a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
				ignore("", "A", ""),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("foo", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			tc("ignore type,target",
				a("foo", "1.2.3.4"),
				// a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
				ignore("", "A", "2.3.4.5"),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("foo", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			tc("ignore target",
				a("foo", "1.2.3.4"),
				// a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
				ignore("", "", "2.3.4.5"),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("foo", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			// Many types:
			tc("ignore manytypes",
				// a("foo", "1.2.3.4"),
				// a("foo", "2.3.4.5"),
				// txt("foo", "simple"),
				// a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
				ignore("", "A,TXT", ""),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("foo", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			// Target with wildcard:
			tc("ignore label,type,target=*",
				a("foo", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				// cname("mail", "ghs.googlehosted.com."),
				ignore("", "CNAME", "*.googlehosted.com."),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("foo", "1.2.3.4"),
				a("foo", "2.3.4.5"),
				txt("foo", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),
		),

		// Same as "main" but with an apex ("@") record.
		testgroup("IGNORE apex",
			tc("Create some records",
				a("@", "1.2.3.4"),
				a("@", "2.3.4.5"),
				txt("@", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			),

			tc("apex label",
				// NB(tlim): This ignores 1 record of a recordSet. This should
				// fail for diff2.ByRecordSet() providers if diff2 is not
				// implemented correctly.
				// a("@", "1.2.3.4"),
				// a("@", "2.3.4.5"),
				// txt("@", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
				ignore("@", "", ""),
				// ignore("", "NS", ""),
				// NB(tlim): .UnsafeIgnore is needed because the NS records
				// that providers injects into zones are treated like input
				// from dnsconfig.js.
			).ExpectNoChanges().UnsafeIgnore(),
			tc("VERIFY PREVIOUS",
				a("@", "1.2.3.4"),
				a("@", "2.3.4.5"),
				txt("@", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			tc("apex label,type",
				// a("@", "1.2.3.4"),
				// a("@", "2.3.4.5"),
				txt("@", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
				ignore("@", "A", ""),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("@", "1.2.3.4"),
				a("@", "2.3.4.5"),
				txt("@", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			tc("apex label,type,target",
				// a("@", "1.2.3.4"),
				a("@", "2.3.4.5"),
				txt("@", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
				ignore("@", "A", "1.2.3.4"),
				// NB(tlim): .UnsafeIgnore is needed because the NS records
				// that providers injects into zones are treated like input
				// from dnsconfig.js.
			).ExpectNoChanges().UnsafeIgnore(),
			tc("VERIFY PREVIOUS",
				a("@", "1.2.3.4"),
				a("@", "2.3.4.5"),
				txt("@", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			tc("apex type",
				// a("@", "1.2.3.4"),
				// a("@", "2.3.4.5"),
				txt("@", "simple"),
				// a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
				ignore("", "A", ""),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("@", "1.2.3.4"),
				a("@", "2.3.4.5"),
				txt("@", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			tc("apex type,target",
				a("@", "1.2.3.4"),
				// a("@", "2.3.4.5"),
				txt("@", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
				ignore("", "A", "2.3.4.5"),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("@", "1.2.3.4"),
				a("@", "2.3.4.5"),
				txt("@", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			tc("apex target",
				a("@", "1.2.3.4"),
				// a("@", "2.3.4.5"),
				txt("@", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
				ignore("", "", "2.3.4.5"),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("@", "1.2.3.4"),
				a("@", "2.3.4.5"),
				txt("@", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			// Many types:
			tc("apex manytypes",
				// a("@", "1.2.3.4"),
				// a("@", "2.3.4.5"),
				// txt("@", "simple"),
				// a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
				ignore("", "A,TXT", ""),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("@", "1.2.3.4"),
				a("@", "2.3.4.5"),
				txt("@", "simple"),
				a("bar", "5.5.5.5"),
				cname("mail", "ghs.googlehosted.com."),
			).ExpectNoChanges(),
		),

		// IGNORE with unsafe notation

		testgroup("IGNORE unsafe",
			tc("Create some records",
				txt("foo", "simple"),
				a("foo", "1.2.3.4"),
				txt("@", "asimple"),
				a("@", "2.2.2.2"),
			),

			tc("ignore unsafe apex",
				txt("foo", "simple"),
				a("foo", "1.2.3.4"),
				txt("@", "asimple"),
				a("@", "2.2.2.2"),
				ignore("@", "", ""),
			).ExpectNoChanges().UnsafeIgnore(),
			tc("VERIFY PREVIOUS",
				txt("foo", "simple"),
				a("foo", "1.2.3.4"),
				txt("@", "asimple"),
				a("@", "2.2.2.2"),
			).ExpectNoChanges(),

			tc("ignore unsafe label",
				txt("foo", "simple"),
				a("foo", "1.2.3.4"),
				txt("@", "asimple"),
				a("@", "2.2.2.2"),
				ignore("foo", "", ""),
			).ExpectNoChanges().UnsafeIgnore(),
			tc("VERIFY PREVIOUS",
				txt("foo", "simple"),
				a("foo", "1.2.3.4"),
				txt("@", "asimple"),
				a("@", "2.2.2.2"),
			).ExpectNoChanges(),
		),

		// IGNORE with wildcards

		testgroup("IGNORE wilds",
			tc("Create some records",
				a("foo.bat", "1.2.3.4"),
				a("foo.bat", "2.3.4.5"),
				txt("foo.bat", "simple"),
				a("bar.bat", "5.5.5.5"),
				cname("mail.bat", "ghs.googlehosted.com."),
			),

			tc("ignore label=foo.*",
				// a("foo.bat", "1.2.3.4"),
				// a("foo.bat", "2.3.4.5"),
				// txt("foo.bat", "simple"),
				a("bar.bat", "5.5.5.5"),
				cname("mail.bat", "ghs.googlehosted.com."),
				ignore("foo.*", "", ""),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("foo.bat", "1.2.3.4"),
				a("foo.bat", "2.3.4.5"),
				txt("foo.bat", "simple"),
				a("bar.bat", "5.5.5.5"),
				cname("mail.bat", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			tc("ignore label=foo.bat,type",
				// a("foo.bat", "1.2.3.4"),
				// a("foo.bat", "2.3.4.5"),
				txt("foo.bat", "simple"),
				// a("bar.bat", "5.5.5.5"),
				cname("mail.bat", "ghs.googlehosted.com."),
				ignore("*.bat", "A", ""),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("foo.bat", "1.2.3.4"),
				a("foo.bat", "2.3.4.5"),
				txt("foo.bat", "simple"),
				a("bar.bat", "5.5.5.5"),
				cname("mail.bat", "ghs.googlehosted.com."),
			).ExpectNoChanges(),

			tc("ignore target=*.domain",
				a("foo.bat", "1.2.3.4"),
				a("foo.bat", "2.3.4.5"),
				txt("foo.bat", "simple"),
				a("bar.bat", "5.5.5.5"),
				// cname("mail.bat", "ghs.googlehosted.com."),
				ignore("", "", "*.googlehosted.com."),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("foo.bat", "1.2.3.4"),
				a("foo.bat", "2.3.4.5"),
				txt("foo.bat", "simple"),
				a("bar.bat", "5.5.5.5"),
				cname("mail.bat", "ghs.googlehosted.com."),
			).ExpectNoChanges(),
		),

		// IGNORE with changes
		testgroup("IGNORE with modify",
			tc("Create some records",
				a("foo", "1.1.1.1"),
				a("foo", "10.10.10.10"),
				aaaa("foo", "2003:dd:d7ff::fe71:aaaa"),
				mx("foo", 10, "aspmx.l.google.com."),
				mx("foo", 20, "alt1.aspmx.l.google.com."),
				a("zzz", "3.3.3.3"),
				a("zzz", "4.4.4.4"),
				aaaa("zzz", "2003:dd:d7ff::fe71:cccc"),
			),

			// ByZone: Change (anywhere)
			tc("IGNORE change ByZone",
				ignore("zzz", "A", ""),
				a("foo", "1.1.1.1"),
				a("foo", "11.11.11.11"), // CHANGE
				aaaa("foo", "2003:dd:d7ff::fe71:aaaa"),
				mx("foo", 10, "aspmx.l.google.com."),
				mx("foo", 20, "alt1.aspmx.l.google.com."),
				// a("zzz", "3.3.3.3"),
				// a("zzz", "4.4.4.4"),
				aaaa("zzz", "2003:dd:d7ff::fe71:cccc"),
			),
			tc("VERIFY PREVIOUS",
				a("foo", "1.1.1.1"),
				a("foo", "11.11.11.11"),
				aaaa("foo", "2003:dd:d7ff::fe71:aaaa"),
				mx("foo", 10, "aspmx.l.google.com."),
				mx("foo", 20, "alt1.aspmx.l.google.com."),
				a("zzz", "3.3.3.3"),
				a("zzz", "4.4.4.4"),
				aaaa("zzz", "2003:dd:d7ff::fe71:cccc"),
			).ExpectNoChanges(),

			// ByLabel: Change within a (name) while we ignore the rest
			tc("IGNORE change ByLabel",
				ignore("foo", "MX", ""),
				a("foo", "1.1.1.1"),
				a("foo", "12.12.12.12"), // CHANGE
				aaaa("foo", "2003:dd:d7ff::fe71:aaaa"),
				// mx("foo", 10, "aspmx.l.google.com."),
				// mx("foo", 20, "alt1.aspmx.l.google.com"),
				a("zzz", "3.3.3.3"),
				a("zzz", "4.4.4.4"),
				aaaa("zzz", "2003:dd:d7ff::fe71:cccc"),
			),
			tc("VERIFY PREVIOUS",
				a("foo", "1.1.1.1"),
				a("foo", "12.12.12.12"),
				aaaa("foo", "2003:dd:d7ff::fe71:aaaa"),
				mx("foo", 10, "aspmx.l.google.com."),
				mx("foo", 20, "alt1.aspmx.l.google.com."),
				a("zzz", "3.3.3.3"),
				a("zzz", "4.4.4.4"),
				aaaa("zzz", "2003:dd:d7ff::fe71:cccc"),
			).ExpectNoChanges(),

			// ByRecordSet: Change within a (name+type) while we ignore the rest
			tc("IGNORE change ByRecordSet",
				ignore("foo", "MX,AAAA", ""),
				a("foo", "1.1.1.1"),
				a("foo", "13.13.13.13"), // CHANGE
				// aaaa("foo", "2003:dd:d7ff::fe71:aaaa"),
				// mx("foo", 10, "aspmx.l.google.com."),
				// mx("foo", 20, "alt1.aspmx.l.google.com"),
				a("zzz", "3.3.3.3"),
				a("zzz", "4.4.4.4"),
				aaaa("zzz", "2003:dd:d7ff::fe71:cccc"),
			),
			tc("VERIFY PREVIOUS",
				a("foo", "1.1.1.1"),
				a("foo", "13.13.13.13"),
				aaaa("foo", "2003:dd:d7ff::fe71:aaaa"),
				mx("foo", 10, "aspmx.l.google.com."),
				mx("foo", 20, "alt1.aspmx.l.google.com."),
				a("zzz", "3.3.3.3"),
				a("zzz", "4.4.4.4"),
				aaaa("zzz", "2003:dd:d7ff::fe71:cccc"),
			).ExpectNoChanges(),

			// Change within a (name+type+data) ("ByRecord")
			tc("IGNORE change ByRecord",
				ignore("foo", "A", "1.1.1.1"),
				// a("foo", "1.1.1.1"),
				a("foo", "14.14.14.14"),
				aaaa("foo", "2003:dd:d7ff::fe71:aaaa"),
				mx("foo", 10, "aspmx.l.google.com."),
				mx("foo", 20, "alt1.aspmx.l.google.com."),
				a("zzz", "3.3.3.3"),
				a("zzz", "4.4.4.4"),
				aaaa("zzz", "2003:dd:d7ff::fe71:cccc"),
			),
			tc("VERIFY PREVIOUS",
				a("foo", "1.1.1.1"),
				a("foo", "14.14.14.14"),
				aaaa("foo", "2003:dd:d7ff::fe71:aaaa"),
				mx("foo", 10, "aspmx.l.google.com."),
				mx("foo", 20, "alt1.aspmx.l.google.com."),
				a("zzz", "3.3.3.3"),
				a("zzz", "4.4.4.4"),
				aaaa("zzz", "2003:dd:d7ff::fe71:cccc"),
			).ExpectNoChanges(),
		),

		// IGNORE repro bug reports

		// https://github.com/StackExchange/dnscontrol/issues/2285
		testgroup("IGNORE_TARGET b2285",
			tc("Create some records",
				cname("foo", "redact1.acm-validations.aws."),
				cname("bar", "redact2.acm-validations.aws."),
			),
			tc("Add a new record - ignoring test.foo.com.",
				ignoreTarget("**.acm-validations.aws.", "CNAME"),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				cname("foo", "redact1.acm-validations.aws."),
				cname("bar", "redact2.acm-validations.aws."),
			).ExpectNoChanges(),
		),

		// https://github.com/StackExchange/dnscontrol/issues/2822
		// Don't send empty updates.
		// A carefully constructed IGNORE() can ignore all the
		// changes. This resulted in the deSEC provider generating an
		// empty upsert, which the API rejected.
		testgroup("IGNORE everything b2822",
			tc("Create some records",
				a("dyndns-city1", "91.42.1.1"),
				a("dyndns-city2", "91.42.1.2"),
				aaaa("dyndns-city1", "2003:dd:d7ff::fe71:ce77"),
				aaaa("dyndns-city2", "2003:dd:d7ff::fe71:ce78"),
			),
			tc("ignore them all",
				a("dyndns-city1", "91.42.1.1"),
				a("dyndns-city2", "91.42.1.2"),
				aaaa("dyndns-city1", "2003:dd:d7ff::fe71:ce77"),
				aaaa("dyndns-city2", "2003:dd:d7ff::fe71:ce78"),
				ignore("dyndns-city1", "A,AAAA", ""),
				ignore("dyndns-city2", "A,AAAA", ""),
			).ExpectNoChanges().UnsafeIgnore(),
			tc("VERIFY PREVIOUS",
				a("dyndns-city1", "91.42.1.1"),
				a("dyndns-city2", "91.42.1.2"),
				aaaa("dyndns-city1", "2003:dd:d7ff::fe71:ce77"),
				aaaa("dyndns-city2", "2003:dd:d7ff::fe71:ce78"),
			).ExpectNoChanges(),
		),

		// https://github.com/StackExchange/dnscontrol/issues/3227
		testgroup("IGNORE w/change b3227",
			tc("Create some records",
				a("testignore", "8.8.8.8"),
				a("testdefined", "9.9.9.9"),
			),
			tc("ignore",
				// a("testignore", "8.8.8.8"),
				a("testdefined", "9.9.9.9"),
				ignore("testignore", "", ""),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("testignore", "8.8.8.8"),
				a("testdefined", "9.9.9.9"),
			).ExpectNoChanges(),

			tc("Verify nothing changed",
				a("testignore", "8.8.8.8"),
				a("testdefined", "9.9.9.9"),
			).ExpectNoChanges(),
			tc("VERIFY PREVIOUS",
				a("testignore", "8.8.8.8"),
				a("testdefined", "9.9.9.9"),
			).ExpectNoChanges(),

			tc("ignore with change",
				// a("testignore", "8.8.8.8"),
				a("testdefined", "2.2.2.2"),
				ignore("testignore", "", ""),
			),
			tc("VERIFY PREVIOUS",
				a("testignore", "8.8.8.8"),
				a("testdefined", "2.2.2.2"),
			).ExpectNoChanges(),
		),

		// OVH features

		testgroup("structured TXT",
			only("OVH"),
			tc("Create TXT",
				txt("spf", "v=spf1 ip4:99.99.99.99 -all"),
				txt("dkim", "v=DKIM1;t=s;p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCzwOUgwGWVIwQG8PBl89O37BdaoqEd/rT6r/Iot4PidtPJkPbVxWRi0mUgduAnsO8zHCz2QKAd5wPe9+l+Stwy6e0h27nAOkI/Edx3qwwWqWSUfwfIBWZG+lrFrhWgSIWCj2/TMkMMzBZJdhVszCzdGQiNPkGvKgjfqW5T0TZt0QIDAQAB"),
				txt("_dmarc", "v=DMARC1; p=none; rua=mailto:dmarc@yourdomain.com")),
			tc("Update TXT",
				txt("spf", "v=spf1 a mx -all"),
				txt("dkim", "v=DKIM1;t=s;p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDk72yk6UML8LGIXFobhvx6UDUntqGzmyie2FLMyrOYk1C7CVYR139VMbO9X1rFvZ8TaPnMCkMbuEGWGgWNc27MLYKfI+wP/SYGjRS98TNl9wXxP8tPfr6id5gks95sEMMaYTu8sctnN6sBOvr4hQ2oipVcBn/oxkrfhqvlcat5gQIDAQAB"),
				txt("_dmarc", "v=DMARC1; p=none; rua=mailto:dmarc@example.com")),
		),

		testgroup("structured TXT as native records",
			only("OVH"),
			tc("Create native OVH records",
				ovhspf("spf", "v=spf1 ip4:99.99.99.99 -all"),
				ovhdkim("dkim", "v=DKIM1;t=s;p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCzwOUgwGWVIwQG8PBl89O37BdaoqEd/rT6r/Iot4PidtPJkPbVxWRi0mUgduAnsO8zHCz2QKAd5wPe9+l+Stwy6e0h27nAOkI/Edx3qwwWqWSUfwfIBWZG+lrFrhWgSIWCj2/TMkMMzBZJdhVszCzdGQiNPkGvKgjfqW5T0TZt0QIDAQAB"),
				ovhdmarc("_dmarc", "v=DMARC1; p=none; rua=mailto:dmarc@yourdomain.com")),
			tc("Update native OVH records",
				ovhspf("spf", "v=spf1 a mx -all"),
				ovhdkim("dkim", "v=DKIM1;t=s;p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDk72yk6UML8LGIXFobhvx6UDUntqGzmyie2FLMyrOYk1C7CVYR139VMbO9X1rFvZ8TaPnMCkMbuEGWGgWNc27MLYKfI+wP/SYGjRS98TNl9wXxP8tPfr6id5gks95sEMMaYTu8sctnN6sBOvr4hQ2oipVcBn/oxkrfhqvlcat5gQIDAQAB"),
				ovhdmarc("_dmarc", "v=DMARC1; p=none; rua=mailto:dmarc@example.com")),
		),

		// PORKBUN features

		testgroup("PORKBUN_URLFWD tests",
			only("PORKBUN"),
			tc("Add a urlfwd", porkbunUrlfwd("urlfwd1", "http://example.com", "", "", "")),
			tc("Update a urlfwd", porkbunUrlfwd("urlfwd1", "http://example.org", "", "", "")),
			tc("Update a urlfwd with metadata", porkbunUrlfwd("urlfwd1", "http://example.org", "permanent", "no", "no")),
		),

		// GCORE features

		testgroup("GCORE metadata tests",
			only("GCORE"),
			tc("Add record with metadata", withMeta(a("@", "1.2.3.4"), map[string]string{
				"gcore_filters":    "geodistance,false;first_n,false,2",
				"gcore_asn":        "1234,2345",
				"gcore_continents": "as,na,an,sa,oc,eu,af",
				"gcore_countries":  "cn,us",
				"gcore_latitude":   "12.34",
				"gcore_longitude":  "67.89",
				"gcore_notes":      "test",
				"gcore_weight":     "12",
				"gcore_ip":         "1.2.3.4",
			})),
			tc("Update record with metadata", withMeta(a("@", "1.2.3.4"), map[string]string{
				"gcore_filters":            "healthcheck,false;geodns,false;first_n,false,3",
				"gcore_failover_protocol":  "HTTP",
				"gcore_failover_port":      "443",
				"gcore_failover_frequency": "30",
				"gcore_failover_timeout":   "10",
				"gcore_failover_method":    "POST",
				"gcore_failover_url":       "/test",
				"gcore_failover_tls":       "false",
				"gcore_failover_regexp":    "",
				"gcore_failover_host":      "example.com",
				"gcore_asn":                "2345,3456",
				"gcore_continents":         "as,na",
				"gcore_countries":          "gb,fr",
				"gcore_latitude":           "12.89",
				"gcore_longitude":          "34.56",
				"gcore_notes":              "test2",
				"gcore_weight":             "34",
				"gcore_ip":                 "4.3.2.1",
			})),
			tc("Delete metadata from record", a("@", "1.2.3.4")),
		),

		// NAMECHEAP features

		testgroup("NAMECHEAP url redirect records",
			only("NAMECHEAP"),
			tc("Create the three types",
				url("unmasked", "https://example.com"),
				url301("permanent", "https://example.com"),
				frame("masked", "https://example.com"),
			),
			tc("VERIFY PREVIOUS",
				url("unmasked", "https://example.com"),
				url301("permanent", "https://example.com"),
				frame("masked", "https://example.com"),
			).ExpectNoChanges(),
		),

		// This MUST be the last test.
		testgroup("final",
			tc("final", txt("final", `TestDNSProviders was successful!`)),
		),

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
