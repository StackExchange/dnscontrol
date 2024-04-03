# Writing new DNS providers

Writing a new DNS provider is a relatively straightforward process.
You essentially need to implement the
[providers.DNSServiceProvider interface.](https://pkg.go.dev/github.com/StackExchange/dnscontrol/providers#DNSServiceProvider)
and the system takes care of the rest.

Please do note that if you submit a new provider you will be
assigned bugs related to the provider in the future (unless
you designate someone else as the maintainer). More details
[here](providers.md).

Please follow the [DNSControl Code Style Guide](styleguide-code.md) and the [DNSControl Documentation Style Guide](styleguide-doc.md).

## Overview

I'll ignore all the small stuff and get to the point.

A typical provider implements 3 methods and DNSControl takes care of the rest:

* GetZoneRecords() -- Download the list of DNS records and return them as a list of RecordConfig structs.
* GetZoneRecordsCorrections() -- Generate a list of corrections.
* GetNameservers() -- Query the API and return the list of parent nameservers.

These three functions are all that's needed for `dnscontrol preview`
and `dnscontrol push`.

The goal of `GetZoneRecords()` is to download all the DNS records,
convert them to `models.RecordConfig` format, and return them as one big list
(`models.Records`).

The goal of `GetZoneRecordsCorrections()` is to return a list of
corrections.  Each correction is a text string describing the change
("Delete CNAME record foo") and a function that, if called, will
make the change (i.e. call the API and delete record foo).  `dnscontrol preview`
simply prints the text strings.  `dnscontrol push` prints the
strings and calls the functions. Because of how Go's functions work,
the function will have everything it needs to make the change.
Pretty cool, eh?

Calculating the difference between existing and desired is difficult. Luckily
the work is done for you.  `GetZoneRecordsCorrections()` calls a a function in
the `pkg/diff2` module that generates a list of changes (usually an ADD,
CHANGE, or DELETE) that can easily be turned into the API calls mentioned
previously.

So, what does all this mean?

It basically means that writing a provider is as simple as writing
code that (1) downloads the existing records, (2) converts each
records into `models.RecordConfig`, (3) write functions that perform
adds, changes, and deletions.

If you are new to Go, there are plenty of providers you can copy
from. In fact, many non-Go programmers
[have learned Go by contributing to DNSControl](https://everythingsysadmin.com/2017/08/go-get-up-to-speed.html).

Now that you understand the general process, here are the details.

## Step 1: General advice

A provider can be a DnsProvider, a Registrar, or both. We recommend
you write the DnsProvider first, release it, and then write the
Registrar if needed.

If you have any questions, please discuss them in the GitHub issue
related to the request for this provider.

This document is constantly being updated.  Please let us know what
was confusing so we can update this document with advice for future
authors (or even better send a PR!).

## Step 2: Pick a base provider

It's a good idea to start by copying a similar provider.

How can you tell a provider is similar?

Each DNS provider's API falls into one of 4 category. Some update one DNS record at a time.
Others, the only change they permit is to upload the entire zone even if only one record changed!
Others are somewhere in between: all records at a label must be updated at once, or all records
in a RecordSet (the label + rType).

In summary, provider APIs basically fall into four general categories:

* Updates are done one record at a time (Record)
* Updates are done one label at a time (Label)
* Updates are done one label+type at a time (RecordSet)
* Updates require the entire zone to be uploaded (Zone).

To determine your provider's category, review your API documentation.

To determine an existing provider's category, see which `diff2.By*()` function is used.

DNSControl provides 4 helper functions that do all the hard work for
you.  As input, they take the existing zone (what was downloaded via
the API) and the desired zone (what is in `dnsconfig.js`).  They
return a list of instructions. Implement handlers for the instructions
and DNSControl is able to perform `dnscontrol push`.

The functions are:

* diff2.ByRecord() -- Updates are done one DNS record at a time. New records are added. Changes and deletes refer to an ID assigned to the record by the provider.
* diff2.ByLabel() -- Updates are done for an entire label. Adds and changes are done by sending one or more records that will appear at that label (i.e. www.example.com). Deletes delete all records at that label.
* diff2.ByRecordSet() -- Similar to ByLabel() but updates are done on the label+type level. If www.example.com has 2 A records and 2 MX records, updates must replace all the A records, or all the MX records, or add records of a different type.
* diff2.ByZone() -- Updates are done by uploading the entire zone every time.

The file `pkg/diff2/diff2.go` has instructions about how to use the diff2 system.

## Step 3: Create the driver skeleton

Create a directory for the provider called `providers/name` where
`name` is all lowercase and represents the commonly-used name for
the service.

The main driver should be called `providers/name/nameProvider.go`.
The API abstraction is usually in a separate file (often called
`api.go`).

Directory names should be consitent.  It should be all lowercase and match the ALLCAPS provider name. Avoid `_`s.

## Step 4: Activate the driver

Edit
[providers/\_all/all.go](https://github.com/StackExchange/dnscontrol/blob/main/providers/_all/all.go).
Add the provider list so DNSControl knows it exists.

## Step 5: Implement

**If you are implementing a DNS Service Provider:**

Implement all the calls in the
[providers.DNSServiceProvider interface](https://pkg.go.dev/github.com/StackExchange/dnscontrol/v4/providers#DNSServiceProvider).

The function `GetDomainCorrections()` is a bit interesting. It returns
a list of corrections to be made. These are in the form of functions
that DNSControl can call to actually make the corrections.

**If you are implementing a DNS Registrar:**

Implement all the calls in the
[providers.Registrar interface](https://pkg.go.dev/github.com/StackExchange/dnscontrol/v4/providers#Registrar).

The function `GetRegistrarCorrections()` returns
a list of corrections to be made. These are in the form of functions
that DNSControl can call to actually make the corrections.

## Step 6: Unit Test

Make sure the existing unit tests work.  Add unit tests for any
complex algorithms in the new code.

Run the unit tests with this command:

    go test ./...

## Step 7: Integration Test

This is the most important kind of testing when adding a new provider.
Integration tests use a test account and a test domain.

{% hint style="danger" %}
All records will be deleted from the test domain!  Use a OTE domain or a real domain that isn't otherwise in use and can be destroyed.
{% endhint %}

* Edit [integrationTest/providers.json](https://github.com/StackExchange/dnscontrol/blob/main/integrationTest/providers.json):
  * Add the `creds.json` info required for this provider in the form of environment variables.

Now you can run the integration tests.

For example, test BIND:

```shell
cd integrationTest              # NOTE: Not needed if already there
export BIND_DOMAIN='example.com'
go test -v -verbose -provider BIND
```

(BIND is a good place to start since it doesn't require API keys.)

This will run the tests on Amazon AWS Route53:

```shell
export R53_DOMAIN='dnscontroltest-r53.com'    # Use a test domain.
export R53_KEY_ID='CHANGE_TO_THE_ID'
export R53_KEY='CHANGE_TO_THE_KEY'
cd integrationTest              # NOTE: Not needed if already there
go test -v -verbose -provider ROUTE53
```

Some useful `go test` flags:

* Run only certain tests using the `-start` and `-end` flags.
  * Rather than running all the tests, run just the tests you want.
  * These flags must be *after* the `-provider FOO` flag.
  * Example: `go test -v -verbose -provider ROUTE53 -start 10 -end 20` run tests 10-20 inclusive.
  * Example: `go test -v -verbose -provider ROUTE53 -start 5 -end 5` runs only test 5.
  * Example: `go test -v -verbose -provider ROUTE53 -start 20` skip the first 19 tests.
  * Example: `go test -v -verbose -provider ROUTE53 -end 20` only run the first 20 tests.
* Slow tests? Add `-timeout n` to increase the timeout for tests
  * `go test` kills the tests after 10 minutes by default.  Some providers need more time.
  * This flag must be *before* the `-verbose` flag.  Usually it is the first flag after `go test`.
  * Example:  `go test -timeout 20m -v -verbose -provider CLOUDFLAREAPI`
* If a test will always fail because the provider doesn't support the feature, you can opt out of the test.  Look at `func makeTests()` in [integrationTest/integration_test.go](https://github.com/StackExchange/dnscontrol/blob/2f65533e1b92c2967229a92a304fff7c14f7f4b6/integrationTest/integration_test.go#L675) for more details.

## Step 8: Manual tests

This is optional.

There is a potential bug in how TXT records are handled. Sadly we haven't found
an automated way to test for this bug.  The manual steps are here in
[documentation/testing-txt-records.md](testing-txt-records.md)

## Step 9: Update docs, CICD and other files

* Edit `README.md`:
  * Add the provider to the bullet list.
* Edit `.github/workflows/pr_test.yml`
  * Add the name of the provider to the PROVIDERS list.
* Edit `documentation/providers.md`:
  * Remove the provider from the `Requested providers` list (near the end of the doc) (if needed).
  * Add the new provider to the [Providers with "contributor support"](providers.md#providers-with-contributor-support) section.
* Edit `documentation/SUMMARY.md`:
  * Add the provider to the "Providers" list.
* Create `documentation/providers/PROVIDERNAME.md`:
  * Use one of the other files in that directory as a base.
* Edit `OWNERS`:
  * Add the directory name and your GitHub username.

{% hint style="success" %}
**Need feedback?** Submit a draft PR!  It's a great way to get early feedback, ask about fixing
a particular integration test, or request feedback.
{% endhint %}

## Step 10: Capabilities

Some DNS providers have features that others do not.  For example some
support the SRV record.  A provider announces what it can do using
the capabilities system.

If a provider doesn't advertise a particular capability, the integration
test system skips the appropriate tests.  Therefore you might want
to initially develop the provider with no particular capabilities
advertised and code until all the integration tests work.  Then
enable capabilities one at a time to finish off the project.

Don't feel obligated to implement everything at once. In fact, we'd
prefer a few small PRs than one big one. Focus on getting the basic
provider working well before adding these extras.

Operational features have names like `providers.CanUseSRV` and
`providers.CanUseAlias`.  The list of optional "capabilities" are
in the file `dnscontrol/providers/providers.go` (look for `CanUseAlias`).

Capabilities are processed early by DNSControl.  For example if a
provider doesn't support SRV records, DNSControl will error out
when parsing `dnscontrol.js` rather than waiting until the API fails
at the very end.

Enable optional capabilities in the `nameProvider.go` file and run
the integration tests to see what works and what doesn't.  Fix any
bugs and repeat, repeat, repeat until you have all the capabilities
you want to implement.

FYI: If a provider's capabilities changes, run `go generate` to update
the documentation.

## Step 11: Automated code tests

Run `go vet` and [`staticcheck`](https://staticcheck.io/) and clean up any errors found.

```shell
go vet ./...
staticcheck ./...
```

Please use `go vet` from the [newest release of Go](https://golang.org/doc/devel/release.html#policy).

golint is deprecated and frozen but it is still useful as it does a few checks that haven't been
re-implemented in staticcheck.
However golink fails on any file that uses generics, so
be prepared to ignore errors about `expected '(', found '[' (and 1 more errors)`

How to install and run [golint](https://github.com/golang/lint):

```shell
go get -u golang.org/x/lint/golint
go install golang.org/x/lint/golint
golint ./...
```

## Step 12: Dependencies

See [documentation/release-engineering.md](release-engineering.md)
for tips about managing modules and checking for outdated
dependencies.

## Step 13: Modify the release regexp

In the repo root, open `.goreleaser.yml` and add the provider to `Provider-specific changes` regexp.

## Step 14: Check your work

These are the things we'll be checking when you submit the PR.  Please try to complete all or as many of these as possible.

1. Run `go generate ./...` to make sure all generated files are fresh.
2. Make sure the following files were created and/or updated:
  * `OWNERS`
  * `README.md`
  * `.github/workflows/pr_test.yml` (The PROVIDERS list)
  * `.goreleaser.yml` (Search for `Provider-specific changes`)
  * `documentation/SUMMARY.md`
  * `documentation/providers.md` (the autogenerated table + the second one; make sure it is removed from the `requested` list)
  * `documentation/providers/`PROVIDERNAME`.md`
  * `integrationTest/providers.json`
    * `providers/_all/all.go`
3. Review the code for style issues, remove debug statements, make sure all exported functions have a comment, and generally tighten up the code.
4. Verify you're using the most recent version of anything you import.  (See [Step 12](#step-12-dependencies))
5. Re-run the [integration test](#step-7-integration-test) one last time.
  * Post the results as a comment to your PR.
6. Re-read the [maintainer's responsibilities](providers.md#providers-with-contributor-support) bullet list.  By submitting a provider you agree to maintain it, respond to bugs, periodically re-run the integration test to verify nothing has broken, and if we don't hear from you for 2 months we may disable the provider.

## Step 15: Submit a PR

At this point you can submit a PR.

Actually you can submit the PR even earlier if you just want feedback,
input, or have questions.  This is just a good stopping place to
submit a PR if you haven't already.

## Step 16: After the PR is merged

1. Close any related GitHub issues.
3. Would you like your provider to be tested automatically as part of every PR?  Sure you would!  Follow the instructions in [Bring-Your-Own-Secrets for automated testing](byo-secrets.md)
