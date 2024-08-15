# Adding DNS Resource Types the "Rdata" way

Terminology:
* RType: A DNS "record type" such as an A, AAAA, CNAME, MX record.
* RC-style: The original way to add new rtypes.
* Rdata-style: The new way to add new rtypes, documented here.

In September 2024 DNSControl gained a new way to implement rtypes called
"Rdata-style".  This can be used to add RFC-standard types such as LOC, as well
as provider-specific types such as Cloudflare's "Single Redirect".

This document explains how RData-style rtypes work and how to add a new record
type using this method.

The old and new styles are both supported.  All new rtypes should use
Rdata-style. There is no need to convert the old rtypes to use RData-style,
though we'll gladly accept PRs that convert existing rtypes to use Rdata.

## Goals

Goals of Rdata-style records:

* **Goal: Make it considerably easier to add a new rtype.**
  * Problem: RC-Style requires writing code in both Go and JavaScript.
  * Solution: Rdata-style requires only writing Go (plus 1 line of JavaScript)
* **Goal: Make testing easier.**
  * Problem: RC-Style has no support for unit testing the JavaScript in
    helpers.js.
  * Solution: Rdata-style only uses Go (with 1 minor exception) and permits the
    use of the standard Go unit testing framework.
* **Goal: Stop increasing the size of models.RecordConfig.**
  * Problem: RC-Style requires each new rtype to add fields to RecordConfig.
    This consumes memory for every RecordConfig instance. For example, the
    DNSKEY rtype added 4 fields, consuming 14 bytes of memory even when the
    RecordConfig is not storying a DNSKEY. (Not to pick on DNSKEY... this was
    the only option at the time!)
  * Solution: RecordConfig now has one field that is a pointer to struct, which
    is the right size for the rtype.
* Goal: Isolate an rtype's implementation in the code base.
  * Problem: RC-Style spreads implementation all over the code base.
  * RData-style: Code is isolated to a specific directory with many exceptions.
    The list of exceptions should shrink over time.

## Conceptual design.

To understand how Rdata-style works, first let's review the old way.

The old way:

RC-style implements a JavaScript function in helpers.js that accepts the
user-input fields, processes them, and makes a JSON version of RecordConfig
which is sent to the Go code for use. It is assumed that the JSON that is
delivered is complete.

For example, `LOC_BUILDER_DD()` is implemented completely in JavaScript. This
is a complex function and, since we lack unit-testing in DNSControl's
JavaScript environment, has no test coverage.

The new way:

In RData-style, the helpers.js function simply collects all the parameters and
delivers them to the Go code verbatium.  A function in Go extracts the
parameters and uses them to build a struct.  models.RecordConfig.Rdata points
to that struct.

For example, `CF_SINGLE_REDIRECT()`'s implementation in helpers.js is one line:

```
var CF_SINGLE_REDIRECT = rawrecordBuilder('CLOUDFLAREAPI_SINGLE_REDIRECT');
```

This creates a function called `CF_SINGLE_REDIRECT()` which users can use in `dnsconfig.js`.

All the remaining code is in `dnscontrol/rtypes/rtype$NAME` (global rtypes)
or `dnscontrol/providers/$PROVIDER/rtypes/rtype$NAME` (provider-specific rtypes).
`$PROVIDER` is the name of the provider, and $NAME is the name of the record.
For example, the Cloudflare Single Redirect type would be in `providers/cloudflare/rtypes/rtypesingleredirect`.

Yes, there is a lot of code outside the rtypes/rtype$NAME directory still.
However we're working on reducing that.

# How to add a new rtype:

## Step 1: Update helpers.js

Edit `pkg/js/helpers.js`

At the end of the file, add a line such as:

```
var CF_SINGLE_REDIRECT = rawrecordBuilder('CLOUDFLAREAPI_SINGLE_REDIRECT');
      ^^^^^                                ^^^^^^^^^^^^^
      function name                        rtype token
```

* function name: This is the name that appears in `dnsconfig.js`.
  * For RFC-standard types this should be the name of the type as it would appear in a zone file. (Example: `A`, `MX`, `LOC`)
  * For provider-specific types, the prefix should be the provider's name or initials (`CF_` for CloudFlare).
  * For pseudo-types that apply to any provider, use your best judgement.
* rtype token: The string that is used in the models.RecordConfig.Type field.
  * For RFC-standard types this should be the name of the type as it would appear in a zone file.
  * For provider-specific types, the prefix should be the provider's name exactly as it is used in `creds.json`.
  * For pseudo-types that apply to any provider, it should be exactly the same as the function name.

## Step X: Implement the rtype's functions

General form:

```
providers/cloudflare/rtypes/rtype$NAME/$NAME.go
```

Example:

```
providers/cloudflare/rtypes/rtypesingleredirect/cfsingleredirect.go
```

Implement:

* `const Name`: Same string as the "rtype token" in helpers.js
* `init()`: Copy verbatim
* Define the struct.  `type $Name struct` where `$Name` is the rtype name in mixed case.
* function `Name`: Copy verbatium
* function `ComputeTarget`: returns the "target field" for the record. For example, an `MX` Record would return the hostname (not the preference number), an `A` record would return the IP address.
* function `ComputeComparableMini`: returns a string representation of all the rtype's fields. This string is used for comparing two records. If there are any differences, the two are not considered the same. This string should be human-readable, since it is used in the output of `dnscontrol preview`.  For example, `MX` would output `50 example.com.`  Note that the label is not included, nor the TTL.
* function `MarshalJSON`: returns a JSON representation of all the rtype's fields. Note that the label is not included, nor the TTL.
* function `FromRawArgs`: Described below.

## Step X: Implement FromRawArgs

This function takes the raw items from helpes.js and builds the struct.

Copy from another rtype.  Here's what the code does:

```
// FromRawArgs creates a Rdata...
// update a RecordConfig using the args (from a
// RawRecord.Args). In other words, use the data from dnsconfig.js's
// rawrecordBuilder to create (actually... update) a models.RecordConfig.
func FromRawArgs(items []any) (*SingleRedirect, error) {
```

The function takes the raw arguments, which arrive as an array of "any"... i.e. they can be any type.

```
  // Pave the arguments.
  if err := rtypecontrol.PaveArgs(items, "iss"); err != nil {
    return nil, err
  }
```

`rtypecontrol.PaveArgs()` takes the raw items and validates them, or fixes them.
The string (in this example, `"iss"`) includes 1 letter for each parameter.

* `i`: uint16: Converts strings, truncates float64s, resizes ints, etc.
* `s`: string: Converts all types to string.

If you desire other types, add them to `pkg/rtypecontrol/pave.go`.

```
  // Unpack the arguments:
  var code = items[0].(uint16)
  if code != 301 && code != 302 {
    return nil, fmt.Errorf("code (%03d) is not 301 or 302", code)
  }
  var when = items[1].(string)
  var then = items[2].(string)
```

You are now certain of the type of each `item[]`. Assign each one to a variable of the appropriate type.
This is also where you can validate the inputs. In this example, `code` must be either 301 or 302.

If you are new to go's "type assertions", here's a simple explanation:

* Go doesn't know what type of data is in `item[]` (they are of type `any`). Therefore, we have to tell Go by adding `.(string)` or `.(uint16)`.  We can trust that these are ZZ

Here's the longer version:

here's how they work:
* Each element of `items[]` is an interface. It can be any type.  Go needs us
  to tell us what type to expect when accessing it. It can't guess for us. This
  isn't Python!
* We tell Go it is a string by referring to it as `items[1].(string)`.  This is
  called a "type assertion" because we are asserting the type, since Go can't
  guess it for us.
* This works great, except there's a catch: We we assert wrong, the code will
  panic.  That's why we have to trust `PaveArgs` to do the right thing.
* Wait!  If Go can't guess the type, how does it know it is wrong?  Well, it
  does know. An interface stores both the value and the type. Therefore it
  can check if we've asserted the wrong type. However, it can't generate code that works for all types. The type assertion tells the code generator what to do.
* The Pave Pattern is something I created for DNSControl to make it easier to work with interfaces.  You won't see it elsewhere.  Most projects make you do all the work yourself.
* To learn more about Go's type assertions and "type switches", a good tutorial is here: [https://rednafi.com/go/type_assertion_vs_type_switches/](https://rednafi.com/go/type_assertion_vs_type_switches/)

```
  // Use the arguments to perfect the record:
  return makeSingleRedirectFromRawRec(code, name, when, then)
}
```

This calls a function that makes the struct (actually a pointer to a struct). For simple record types there's no need to make this a separate function.

## Step X: ConvertRawRecords

Edit models/rawrecord.go

In the function `ConvertRawRecords()`, add to the switch statement a case for the new type.

Here's an example. Change "foo" to the name of your type.

```
      case rtypefoo.Name:
        rdata, error := rtypefoo.FromRawArgs(args, label)
        if error != nil {
          return err
        }
        rec.Seal(dc.Name, label, rdata)
```

## Step X: update casts.go

This will be automated some day, but in the meanwhile this is done manually.

Edit `models/casts.go`

Add the rtype's module to the imports list.

Add the rtype's `As*()` function. For example, if you are adding an rtype FOO, add a function`AsFOO()`.

Follow the examples.

```
import (
  "github.com/StackExchange/dnscontrol/v4/providers/cloudflare/rtypes/rtypefoo"
)

func (rc *RecordConfig) AsFOO() *rtypefoo.FOO {
  return rc.Rdata.(*rtypefoo.FOO)
}
```

## Step X: update create.go

This will be automated some day, but in the meanwhile this is done manually.

Edit `pkg/create/create.go`

Add the rtype's module to the imports list.

Add the rtype's `Foo()` function. For example, if you are adding an rtype FOO, add a function`FOO()`.

Follow the examples. It should be exactly the same as `SingleRedirect()` with `SingleRedirect` changed to `FOO`.

## Step X: Add this to a provider

## Step X: Add integration etsts


## Step 2: Add a capability for the record

You'll need to mark which providers support this record type. If BIND supports this record type, add support
to bind first.  This is the easiest provider to update.  Otherwise choose another provider.

-   Add the capability to the file `dnscontrol/providers/capabilities.go` (look for `CanUseAlias` and add
    it to the end of the list.)
-   Run stringer to auto-update the file `dnscontrol/providers/capability_string.go`

Install stringer:

```
go install golang.org/x/tools/cmd/stringer@latest
```

Run stringer:

```shell
cd providers
go generate
```

-   Add this feature to the feature matrix in `dnscontrol/build/generate/featureMatrix.go`. Add it to the variable `matrix` maintaining alphabetical ordering, which should look like this:

    {% code title="dnscontrol/build/generate/featureMatrix.go" %}
    ```diff
    func matrixData() *FeatureMatrix {
        const (
            ...
            DomainModifierCaa    = "[`CAA`](language-reference/domain-modifiers/CAA.md)"
    +       DomainModifierFoo    = "[`FOO`](language-reference/domain-modifiers/FOO.md)"
            DomainModifierLoc    = "[`LOC`](language-reference/domain-modifiers/LOC.md)"
            ...
        )
        matrix := &FeatureMatrix{
            Providers: map[string]FeatureMap{},
            Features: []string{
                ...
                DomainModifierCaa,
    +           DomainModifierFoo,
                DomainModifierLoc,
                ...
            },
        }
    ```
    {% endcode %}

    then add it later in the file with a `setCapability()` statement, which should look like this:

    {% code title="dnscontrol/build/generate/featureMatrix.go" %}
    ```diff
    ...
    +       setCapability(
    +           DomainModifierFoo,
    +           providers.CanUseFOO,
    +       )
    ...
    ```
    {% endcode %}

-   Add the capability to the list of features that zones are validated
    against (i.e. if you want DNSControl to report an error if this
    feature is used with a DNS provider that doesn't support it). That's
    in the `checkProviderCapabilities` function in
    `pkg/normalize/validate.go`. It should look like this:

    {% code title="pkg/normalize/validate.go" %}
    ```diff
    var providerCapabilityChecks = []pairTypeCapability{
    ...
    +   capabilityCheck("FOO", providers.CanUseFOO),
    ...
    ```
    {% endcode %}

-   Mark the `bind` provider as supporting this record type by updating `dnscontrol/providers/bind/bindProvider.go` (look for `providers.CanUse` and you'll see what to do).

DNSControl will warn/error if this new record is used with a
provider that does not support the capability.

-   Add the capability to the validations in `pkg/normalize/validate.go`
    by adding it to `providerCapabilityChecks`
-   Some capabilities can't be tested for. If
    such testing can't be done, add it to the whitelist in function
    `TestCapabilitiesAreFiltered` in
    `pkg/normalize/capabilities_test.go`

If the capabilities testing is not configured correctly, `go test ./...`
will report something like the `MISSING` message below. In this
example we removed `providers.CanUseCAA` from the
`providerCapabilityChecks` list.

```text
--- FAIL: TestCapabilitiesAreFiltered (0.00s)
    capabilities_test.go:66: ok: providers.CanUseAlias (0) is checked for with "ALIAS"
    capabilities_test.go:68: MISSING: providers.CanUseCAA (1) is not checked by checkProviderCapabilities
    capabilities_test.go:66: ok: providers.CanUseNAPTR (3) is checked for with "NAPTR"
```

## Step 4: Search for `#rtype_variations`

Anywhere a `rtype` requires special handling has been marked with a
comment that includes the string `#rtype_variations`. Search for
this string and add your new type to this code.

## Step 5: Add a `parse_tests` test case

Add at least one test case to the `pkg/js/parse_tests` directory.
Test `013-mx.js` is a very simple one and is good for cloning.
See also `017-txt.js`.

Run these tests via:

```shell
cd pkg/js/
go test ./...
```

If this works, then you know the `dnsconfig.js` and `helpers.js`
code is working correctly.

As you debug, if there are places that haven't been marked
`#rtype_variations` that should be, add such a comment.
Every time you do this, an angel gets its wings.

The tests also verify that for every "capability" there is a
validation. This is explained in Step 2 (search for
`TestCapabilitiesAreFiltered` or `MISSING`)

## Step 6: Add an `integrationTest` test case

Add at least one test case to the `integrationTest/integration_test.go`
file. Look for `func makeTests` and add the test to the end of this
list.

Each `testgroup()` is a named list of tests.

{% code title="integration_test.go" lineNumbers="true" %}
```go
testgroup("MX",
  tc("MX record", mx("@", 5, "foo.com.")),
  tc("Change MX pref", mx("@", 10, "foo.com.")),
  tc("MX record",
      mx("@", 10, "foo.com."),
      mx("@", 20, "bar.com."),
  ),
)
```
{% endcode %}

Line 1: `testgroup()` gives a name to a group of tests. It also tells
the system to delete all records for this domain so that the tests
begin with a blank slate.

Line 2:
Each `tc()` encodes all the records of a zone. The test framework
will try to do the smallest changes to bring the zone up to date.
In this case, we know the zone is empty, so this will add one MX
record.

Line 3: In this example, we just change one field of an existing
record. To get to this configuration, the provider will have to
either change the priority on an existing record, or delete the old
record and insert a new one. Either way, this test case assures us
that the diff'ing functionality is working properly.

If you look at the tests for `CAA`, it inserts a few records then
attempts to modify each field of a record one at a time. This test
was useful because it turns out we hadn't written the code to
properly see a change in priority. We fixed this bug before the
code made it into production.

Line 4: In this example, the next zone adds a second MX record.
To get to this configuration, the provider will add an
additional MX record to the same label. New tests don't need to do
this kind of test because we're pretty sure that that part of the diffing
engine works fine. It is here as an example.

Also notice that some tests include `requires()`, `not()` and `only()`
statements. This is how we restrict tests to certain providers.
These options must be listed first in a `testgroup`. More details are
in the source code.

To run the integration test with the BIND provider:

```shell
cd integrationTest              # NOTE: Not needed if already in that subdirectory
go test -v -verbose -provider BIND
```

Once the code works for BIND, consider submitting a PR at this point.
(The earlier you submit a PR, the earlier we can provide feedback.)

If you find places that haven't been marked
`#rtype_variations` but should be, please add that comment.
Every time you fail to do this, God kills a cute little kitten.
Please do it for the kittens.

## Step 7: Support more providers

Now add support in other providers. Add the `providers.CanUse...`
flag to the provider and re-run the integration tests:

For example, this will run the tests on Amazon AWS Route53:

```shell
export R53_DOMAIN=dnscontroltest-r53.com  # Use a test domain.
export R53_KEY_ID=CHANGE_TO_THE_ID
export R53_KEY='CHANGE_TO_THE_KEY'
cd integrationTest              # NOTE: Not needed if already in that subdirectory
go test -v -verbose -provider ROUTE53
```

The test should reveal any bugs. Keep iterating between fixing the
code and running the tests. When the tests all work, you are done.
(Well, you might want to clean up some code a bit, but at least you
know that everything is working.)

If you find bugs that aren't covered by the tests, please please
please add a test that demonstrates the bug (then fix the bug, of
course). This
will help all future contributors. If you need help with adding
tests, please ask!

## Step 8: Write documentation

Add a new Markdown file to `documentation/language-reference/domain-modifiers`. Copy an existing file (`CNAME.md` is a good example). The section between the lines of `---` is called the front matter and it has the following keys:

-   `name`: The name of the record. This should match the file name and the name of the record in `helpers.js`.
-   `parameters`: A list of parameter names, in order. Feel free to use spaces in the name if necessary. Your last parameter should be `modifiers...` to allow arbitrary modifiers like `TTL` to be applied to your record.
-   `parameter_types`: an object with parameter names as keys and TypeScript type names as values. Check out existing record documentation if you’re not sure to put for a parameter. Note that this isn’t displayed on the website, it’s only used to generate the `.d.ts` file.

The rest of the file is the documentation. You can use Markdown syntax to format the text.

Add the new file `FOO.md` to the documentation table of contents [`documentation/SUMMARY.md`](SUMMARY.md#domain-modifiers), and/or to the [`Service Provider specific`](SUMMARY.md#service-provider-specific) section if you made a record specific to a provider, and to the [`Record Modifiers`](SUMMARY.md#record-modifiers) section if you created any `*_BUILDER` or `*_HELPER` or similar functions for the new record type:

{% code title="documentation/SUMMARY.md" %}
```diff
...
* Domain Modifiers
...
    * [DnsProvider](language-reference/domain-modifiers/DnsProvider.md)
+   * [FOO](language-reference/domain-modifiers/FOO.md)
    * [FRAME](language-reference/domain-modifiers/FRAME.md)
...
    * Service Provider specific
...
        * ClouDNS
            * [CLOUDNS_WR](language-reference/domain-modifiers/CLOUDNS_WR.md)
+       * ASDF
+           * [ASDF_NINJA](language-reference/domain-modifiers/ASDF_NINJA.md)
        * NS1
            * [NS1_URLFWD](language-reference/domain-modifiers/NS1_URLFWD.md)
...
* Record Modifiers
...
    * [DMARC_BUILDER](language-reference/domain-modifiers/DMARC_BUILDER.md)
+   * [FOO_HELPER](language-reference/record-modifiers/FOO_HELPER.md)
    * [SPF_BUILDER](language-reference/domain-modifiers/SPF_BUILDER.md)
...
```
{% endcode %}

## Step 9: "go generate"

Re-generate the documentation:

```shell
go generate ./...
```

This will regenerate things like the table of which providers have which features and the `dnscontrol.d.ts` file.
