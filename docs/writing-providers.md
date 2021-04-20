---
layout: default
title: Writing new DNS providers
---

# Writing new DNS providers

Writing a new DNS provider is a relatively straightforward process.
You essentially need to implement the
[providers.DNSServiceProvider interface.](https://godoc.org/github.com/StackExchange/dnscontrol/providers#DNSServiceProvider)
and the system takes care of the rest.

Please do note that if you submit a new provider you will be
assigned bugs related to the provider in the future (unless
you designate someone else as the maintainer). More details
[here](provider-list.md).

## Overview

I'll ignore all the small stuff and get to the point.

A provider's `GetDomainCorrections()` function is the workhorse
of the provider.  It is what gets called by `dnscontrol preview`
and `dnscontrol push`.

How does a provider's `GetDomainCorrections()` function work?

The goal of `GetDomainCorrections()` is to return a list of
corrections.  Each correction is a text string describing the change
("Delete CNAME record foo") and a function that, if called, will
make the change (i.e. call the API and delete record foo).  Preview
mode simply prints the text strings.  `dnscontrol push` prints the
strings and calls the functions. Because of how Go's functions work,
the function will have everything it needs to make the change.
Pretty cool, eh?

So how does `GetDomainCorrections()` work?

First, some terminology: The DNS records specified in the dnsconfig.js
file are called the "desired" records. The DNS records stored at
the DNS service provider are called the "existing" records.

Every provider does the same basic process.  The function
`GetDomainCorrections()` is called with a list of the desired DNS
records (`dc.Records`).  It then contacts the provider's API and
gathers the existing records. It converts the existing records into
a list of `*models.RecordConfig`.

Now that it has the desired and existing records in the appropriate
format, `differ.IncrementalDiff(existingRecords)` is called and
does all the hard work of understanding the DNS records and figuring
out what changes need to be made.  It generates lists of adds,
deletes, and changes.

`GetDomainCorrections()` then generates the list of `models.Corrections()`
and returns.  DNSControl takes care of the rest.

So, what does all this mean?

It basically means that writing a provider is as simple as writing
code that (1) downloads the existing records, (2) converts each
records into `models.RecordConfig`, (3) write functions that perform
adds, changes, and deletions.

If you are new to Go, there are plenty of providers you can copy
from. In fact, many non-Go programmers
[have learned Go by contributing to DNSControl](https://everythingsysadmin.com/2017/08/go-get-up-to-speed.html).

Oh, and what if the API simply requires that the entire zonefile be uploaded
every time?  We still generate the text descriptions of the changes (so that
`dnscontrol preview` looks nice) but the functions are just no-ops, except
for one that uploads the new zonefile.

Now that you understand the general process, here are the details.

## Step 1: General advice

A provider can be a DnsProvider, a Registrar, or both. We recommend
you write the DnsProvider first, release it, and then write the
Registrar if needed.

If you have any questions, please discuss them in the Github issue
related to the request for this provider. Please let us know what
was confusing so we can update this document with advice for future
authors (or even better, update [this document](https://github.com/StackExchange/dnscontrol/blob/master/docs/writing-providers.md)
yourself.)


## Step 2: Pick a base provider

Pick a similar provider as your base.  Providers basically fall
into three general categories:

* **zone:** The API requires you to upload the entire zone every time. (BIND, NAMECHEAP).
* **incremental-record:** The API lets you add/change/delete individual DNS records. (ACTIVEDIR, CLOUDFLARE, DNSIMPLE, NAMEDOTCOM, GCLOUD, HEXONET)
* **incremental-label:** Like incremental-record, but if there are
  multiple records on a label (for example, example www.example.com
has A and MX records), you have to replace all the records at that
label. (GANDI_V5)
* **incremental-label-type:** Like incremental-record, but updates to any records at a label have to be done by type.  For example, if a label (www.example.com) has many A and MX records, even the smallest change to one of the A records requires replacing all the A records. Any changes to the MX records requires replacing all the MX records.  If an A record is converted to a CNAME, one must remove all the A records in one call, and add the CNAME record with another call.  This is deceptively difficult to get right; if you have the choice between incremental-label-type and incremental-label, pick incremental-label. (DESEC, ROUTE53)

All providers use the "diff" module to detect differences. It takes
two zones and returns records that are unchanged, created, deleted,
and modified.
The zone providers use the
information to print a human-readable list of what is being changed,
but upload the entire new zone.
The incremental providers use the differences to
update individual records or recordsets.


## Step 3: Create the driver skeleton

Create a directory for the provider called `providers/name` where
`name` is all lowercase and represents the commonly-used name for
the service.

The main driver should be called `providers/name/nameProvider.go`.
The API abstraction is usually in a separate file (often called
`api.go`).


## Step 4: Activate the driver

Edit
[providers/\_all/all.go](https://github.com/StackExchange/dnscontrol/blob/master/providers/_all/all.go).
Add the provider list so DNSControl knows it exists.

## Step 5: Implement

Implement all the calls in
[providers.DNSServiceProvider interface.](https://godoc.org/github.com/StackExchange/dnscontrol/providers#DNSServiceProvider).

The function `GetDomainCorrections` is a bit interesting. It returns
a list of corrections to be made. These are in the form of functions
that DNSControl can call to actually make the corrections.

## Step 6: Unit Test

Make sure the existing unit tests work.  Add unit tests for any
complex algorithms in the new code.

Run the unit tests with this command:

    cd dnscontrol
    go test ./...


## Step 7: Integration Test

This is the most important kind of testing when adding a new provider.
Integration tests use a test account and a real domain.

* Edit [integrationTest/providers.json](https://github.com/StackExchange/dnscontrol/blob/master/integrationTest/providers.json): Add the creds.json info required for this provider.

For example, this will run the tests using BIND:

```
cd dnscontrol/integrationTest
go test -v -verbose -provider BIND
```

(BIND is a good place to  start since it doesn't require any API keys.)

This will run the tests on Amazon AWS Route53:

```
export R53_DOMAIN=dnscontroltest-r53.com  # Use a test domain.
export R53_KEY_ID='CHANGE_TO_THE_ID'
export R53_KEY='CHANGE_TO_THE_KEY'
go test -v -verbose -provider ROUTE53
```

## Step 8: Update docs

* Edit [README.md](https://github.com/StackExchange/dnscontrol): Add the provider to the bullet list.
* Edit [docs/provider-list.md](https://github.com/StackExchange/dnscontrol/blob/master/docs/provider-list.md): Add the provider to the provider list.
* Create `docs/_providers/PROVIDERNAME.md`: Use one of the other files in that directory as a base.
* Edit [OWNERS](https://github.com/StackExchange/dnscontrol/blob/master/OWNERS): Add the directory name and your github id.

## Step 9: Submit a PR

At this point you can submit a PR.

Actually you can submit the PR even earlier if you just want feedback,
input, or have questions.  This is just a good stopping place to
submit a PR if you haven't already.


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
when parsing dnscontrol.js rather than waiting until the API fails
at the very end.

Enable optional capabilities in the nameProvider.go file and run
the integration tests to see what works and what doesn't.  Fix any
bugs and repeat, repeat, repeat until you have all the capabilities
you want to implement.

FYI: If a provider's capabilities changes, run `go generate` to update
the documentation.


## Step 11: Clean up

Run "go vet" and "golint" and clean up any errors found.

```
go vet ./...
golint ./...
```

Please use `go vet` from the [newest release of Go](https://golang.org/doc/devel/release.html#policy).

If [golint](https://github.com/golang/lint) isn't installed on your machine:

```
go get -u golang.org/x/lint/golint
```


## Step 12: Dependencies

See
[docs/release-engineering.md](https://github.com/StackExchange/dnscontrol/blob/master/docs/release-engineering.md)
for tips about managing modules and checking for outdated
dependencies.


## Step 13: Check your work.

Here are some last-minute things to check before you submit your PR.

1. Run "go generate" to make sure all generated files are fresh.
2. Make sure all appropriate documentation is current. (See Step 8)
3. Check that dependencies are current (See Step 12)
4. Re-run the integration test one last time (See Step 7)

## Step 14: After the PR is merged

1. Remove the "provider-request" label from the PR.
2. Verify that [docs/provider-list.md](https://github.com/StackExchange/dnscontrol/blob/master/docs/provider-list.md) no longer shows the provider as "requested"
