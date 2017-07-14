---
layout: default
---

# Writing new DNS providers

Writing a new DNS provider is a relatively straightforward process. You essentially need to implement the [providers.DNSServiceProvider interface.](https://godoc.org/github.com/StackExchange/dnscontrol/providers#DNSServiceProvider) and the system takes care of the rest.

Use another provider as the basis for yours. Pick the provider that has the most similar
API. Some APIs update one DNS record at a time (cloudflare, namedotcom, activedir), some APIs update
all the DNS records on a particular label, some APIs require you to upload the entire
zone every time (bind, gandi).

You'll notice that most providers use the "diff" module to detect differences. It takes
two zones and returns records that are unchanged, created, deleted, and modified. Even
if the API simply requires the entire zone to be uploaded each time, we prefer to list
the specific changes so the user knows exactly what changed.

Here are some tips:

* A provider can be a DnsProvider, a Registrar, or both. We recommend you write the DnsProvider first, release it, and then write the Registrar if needed.
* Create a directory for the provider called `providers/name` where `name` is all lowercase and represents the commonly-used name for the service.
* The main driver should be called `providers/name/nameProvider.go`.  The API abstraction is usually in a separate file (often called `api.go`).
* List the provider in `providers/_all/all.go` so DNSControl knows it exists.
* Implement all the calls in [providers.DNSServiceProvider interface.](https://godoc.org/github.com/StackExchange/dnscontrol/providers#DNSServiceProvider).  The function `GetDomainCorrections` is a bit interesting. It returns a list of corrections to be made. These are in the form of functions that DNSControl can call to actually make the corrections.
* If you have any questions, please dicuss them in the Github issue related to the request for this provider. Please let us know what was confusing so we can update this document with advice for future authors (or feel free to update [this document](https://github.com/StackExchange/dnscontrol/blob/master/docs/writing-providers.md) yourself!).
* Add the provider to the provider list: [docs/provider-list.html](https://github.com/StackExchange/dnscontrol/blob/master/docs/provider-list.html).
* Add the provider to the README: [README.md](https://github.com/StackExchange/dnscontrol)

## Documentation

Please add a page to the docs folder for your provider, and add it to the list in the main project readme.

## Vendoring Dependencies

If your provider depends on other go packages, then you must vendor them. To do this, use [govendor](https://github.com/kardianos/govendor). 

```
go get github.com/kardianos/govendor
govendor add +e
```

is usually sufficient.
