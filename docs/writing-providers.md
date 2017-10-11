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

## Step 1: General advice

A provider can be a DnsProvider, a Registrar, or both. We recommend
you write the DnsProvider first, release it, and then write the
Registrar if needed.

If you have any questions, please dicuss them in the Github issue
related to the request for this provider. Please let us know what
was confusing so we can update this document with advice for future
authors (or even better, update [this
document](https://github.com/StackExchange/dnscontrol/blob/master/docs/writing-providers.md)
yourself.)


## Step 2: Pick a base provider

Pick a similar provider as your base.  Providers basically fall
into three general categories:

* **zone:** The API requires you to upload the entire zone every time. (BIND, GANDI).
* **incremental-record:** The API lets you add/change/delete individual DNS records. (ACTIVEDIR, CLOUDFLARE, NAMEDOTCOM, GCLOUD, ROUTE53)
* **incremental-label:** Similar to incremental, but the API requires you to update all the records related to a particular label each time. For example, if a label (www.example.com) has an A and MX record, any change requires replacing all the records for that label.

TODO: Categorize DNSIMPLE, NAMECHEAP

All providers use the "diff" module to detect differences. It takes
two zones and returns records that are unchanged, created, deleted,
and modified. The incremental providers use the differences to
update individual records or recordsets. The zone providers use the
information to print a human-readable list of what is being changed,
but upload the entire new zone.


## Step 3: Create the driver skeleton

Create a directory for the provider called `providers/name` where
`name` is all lowercase and represents the commonly-used name for
the service.

The main driver should be called `providers/name/nameProvider.go`.
The API abstraction is usually in a separate file (often called
`api.go`).


## Step 4: Activate the driver

Edit
[providers/_all/all.go](https://github.com/StackExchange/dnscontrol/blob/master/providers/_all/all.go).
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
export R53_KEY_ID=CHANGE_TO_THE_ID
export R53_KEY='CHANGE_TO_THE_KEY'
go test -v -verbose -provider ROUTE53
```

## Step 5: Update docs

* Edit [README.md](https://github.com/StackExchange/dnscontrol): Add the provider to the bullet list.
* Edit [docs/provider-list.md](https://github.com/StackExchange/dnscontrol/blob/master/docs/provider-list.md): Add the provider to the provider list.
* Create `docs/_providers/PROVIDERNAME.md`: Use one of the other files in that directory as a base.


## Step 6: Submit a PR

At this point you can submit a PR.

Actually you can submit the PR even earlier if you just want feedback,
input, or have questions.  This is just a good stopping place to
submit a PR. At a minimum a new provider should pass all the
integration tests. Everything else is a bonus.


## Step 7: Capabilities

The last step is to add any optional provider capabilities. You can
submit these as a separate PR once the main provider is working.
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
bugs and repeat.


## Vendoring Dependencies

If your provider depends on other go packages, then you must vendor them. To do this, use [govendor](https://github.com/kardianos/govendor).  A command like this is usually suffient:

```
go get github.com/kardianos/govendor
govendor add +e
```
