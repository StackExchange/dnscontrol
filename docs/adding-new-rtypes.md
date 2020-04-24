---
layout: default
title: Creating new DNS Resource Types (rtypes)
---

# Creating new DNS Resource Types (rtypes)

Everyone is familiar with A, AAAA, CNAME, NS and other Rtypes.
However there are new record types being added all the time (possibly
too many).  Each new record type requires special handling by
DNSControl.

If a record simply has a single "target", then there is little to
do because it is handled similarly to A, CNAME, and so on.  However
if there are multiple fields within the record you have more work
to do.

Our general philosophy is:

* Internally the individual fields of a record are kept separate. If a particular provider combines them into one big string, that kind of thing is done in the provider code at the end of the food chain.  For example, an MX record has a Target (`aspmx.l.google.com.`) and a preference (`10`).  Some systems combine this into one string (`10 aspmx.l.google.com.`).  We keep the two values separate in `RecordConfig` and leave it up to the individual providers to merge them when required. An earlier implementation kept everything combined and we found ourselves constantly parsing and re-parsing the target. It was inefficient and lead to many bugs.
* Anywhere we have a special case for a particular Rtype, we use a `switch` statement and have a `case` for every single record type, usually with a `default:` case that calls `panic()`. This way developers adding a new record type will quickly find where they need to add code (the panic will tell them where).  Before we did this, missing implementation code would go unnoticed for months.
* Keep things alphabetical. If you are adding your record type to a case statement, function library, or whatever, please list it alphabetically along with the others when possible.

## Step 1: Update `RecordConfig` in `models/dns.go`

If the record has any unique fields, add them to `RecordConfig`.
The field name should be the record type, then the field name as
used in `github.com/miekg/dns/types.go`. For example, the `CAA`
record has a field called `Flag`, therefore the field name in
`RecordConfig` is CaaFlag (not `CaaFlags` or `CAAFlags`).

Here are some examples:

```
type RecordConfig struct {
  ...
  MxPreference uint16            `json:"mxpreference,omitempty"` // FIXME(tlim): Rename to MxPreference
  SrvPriority  uint16            `json:"srvpriority,omitempty"`
  SrvWeight    uint16            `json:"srvweight,omitempty"`
  SrvPort      uint16            `json:"srvport,omitempty"`
  CaaTag       string            `json:"caatag,omitempty"`
  CaaFlag      uint8             `json:"caaflag,omitempty"`
  ...
}
```

## Step 2: Add a capability for the record

You'll need to mark which providers support this record type.  The
initial PR should implement this record for the `bind` provider at
a minimum, unless this is a fake or pseudo-type that only a particular
provider supports.

* Add the capability to the file `dnscontrol/providers/capabilities.go` (look for `CanUseAlias` and add
it to the end of the list.)
* Add this feature to the feature matrix in `dnscontrol/build/generate/featureMatrix.go` (Add it to the variable `matrix` then add it later in the file with a `setCap()` statement.
* Add the capability to the list of features that zones are validated
  against (i.e. if you want dnscontrol to report an error if this
  feature is used with a DNS provider that doesn't support it). That's
  in the `checkProviderCapabilities` function in
  `pkg/normalize/validate.go`.
* Mark the `bind` provider as supporting this record type by updating `dnscontrol/providers/bind/bindProvider.go` (look for `providers.CanUse` and you'll see what to do).

DNSControl will warn/error if this new record is used with a
provider that does not support the capability.

* Add the capability to the validations in `pkg/normalize/validate.go`
  by adding it to `providerCapabilityChecks`
* Some capabilities can't be tested for, such as `CanUseTXTMulti`.  If
  such testing can't be done, add it to the whitelist in function
  `TestCapabilitiesAreFiltered` in
  `pkg/normalize/capabilities_test.go`

If the capabilities testing is not configured correctly, `go test ./...`
will report something like the `MISSING` message below. In this
example we removed `providers.CanUseCAA` from the
`providerCapabilityChecks` list.

```
--- FAIL: TestCapabilitiesAreFiltered (0.00s)
    capabilities_test.go:66: ok: providers.CanUseAlias (0) is checked for with "ALIAS"
    capabilities_test.go:68: MISSING: providers.CanUseCAA (1) is not checked by checkProviderCapabilities
    capabilities_test.go:66: ok: providers.CanUseNAPTR (3) is checked for with "NAPTR"
```

## Step 3: Add a helper function

Add a function to `pkg/js/helpers.js` for the new record type.  This
is the JavaScript file that defines `dnsconfig.js`'s functions like
`A()` and `MX()`.  Look at the definition of A, MX and CAA for good
examples to use as a base.

Please add the function alphabetically with the others. Also, please run
[prettier](https://github.com/prettier/prettier) on the file to ensure
your code conforms to our coding standard:

    npm install prettier
    node_modules/.bin/prettier --write pkg/js/helpers.js

FYI: If you change `pkg/js/helpers.js`, run `go generate` to update `pkg/js/static.go`.

## Step 4: Search for `#rtype_variations`

Anywhere a rtype requires special handling has been marked with a
comment that includes the string `#rtype_variations`.  Search for
this string and add your new type to this code.

## Step 5: Add a `parse_tests` test case.

Add at least one test case to the `pkg/js/parse_tests` directory.
Test `013-mx.js` is a very simple one and is good for cloning.

Run these tests via:

    cd dnscontrol/pkg/js
    go test ./...

If this works, then you know the `dnsconfig.js` and `helpers.js`
code is working correctly.

As you debug, if there are places that haven't been marked
`#rtype_variations` that should be, add such a comment.
Every time you do this, an angel gets its wings.

The tests also verify that for every "capability" there is a
validation. This is explained in Step 2 (search for
`TestCapabilitiesAreFiltered` or `MISSING`)

## Step 6: Add an `integrationTest` test case.

Add at least one test case to the `integrationTest/integration_test.go`
file. Look for `var tests =` and add the test to the end of this
list.

Each entry in the list is a new state.  For example:

```
  // MX
  tc("Empty"),                                    <<< 1
  tc("MX record", mx("@", 5, "foo.com.")),        <<< 2
  tc("Change MX pref", mx("@", 10, "foo.com.")),  <<< 3
```

Line 1: An `tc()` entry with no records (just a comment). The test
system will delete all records from the domain to make the domain
match this empty configuration. This creates a "clean slate"
situation.

Line 2: A `tc()` entry with 1 record.  To get to this state, the
provider will have to add the record. If this works, basic functionality
for the MX record type has been achieved.

Line 3: A `tc()` entry with 1 record, with a different priority.
To get to this state, the provider will have to either change the
priority on an existing record, or delete the old record and insert
a new one. Either way, this test case assures us that the diff'ing
functionality is working properly.

If you look at the tests for `CAA`, it inserts a few records then
attempts to modify each field of a record one at a time.  This test
was useful because it turns out we hadn't written the code to
properly see a change in priority. We fixed this bug before the
code made it into production.

Also notice that some tests include `.IfHasCapability()`. This
limits the test to providers with certain capabilities.  You'll
want to use this feature so that the tests only run on providers
that support your new record type.

To run the integration test with the BIND provider:

    cd dnscontrol/integrationTest
    go test -v -verbose -provider BIND

Once the code works for BIND, consider submitting a PR at this point.

As you debug, if there are places that haven't been marked
`#rtype_variations` that should be, add such a comment.
If you fail to do this, God kills a cute little kitten.

## Step 7: Support more providers

Now add support other providers.  Add the `providers.CanUse...`
flag to the provider and re-run the integration tests:

For example, this will run the tests on Amazon AWS Route53:

    export R53_DOMAIN=dnscontroltest-r53.com  # Use a test domain.
    export R53_KEY_ID=CHANGE_TO_THE_ID
    export R53_KEY='CHANGE_TO_THE_KEY'
    go test -v -verbose -provider ROUTE53

The test should reveal any bugs. Keep iterating between fixing the
code and running the tests. When the tests all work, you are done.
(Well, you might want to clean up some code a bit, but at least you
know that everything is working.)

If you find bugs that aren't covered by the tests, please please
please add a test that demonstrates the bug THEN fix the bug. This
will help all future contributors. If you need help with adding
tests, please ask!
