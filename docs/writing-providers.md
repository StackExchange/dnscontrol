---
layout: default
---

# Writing new DNS providers

Writing a new DNS provider is a relatively straightforward process. You essentially need to implement the [providers.DNSServiceProvider interface.](https://godoc.org/github.com/StackExchange/dnscontrol/providers#DNSServiceProvider)

...

More info to follow soon.

## Documentation

Please add a page to the docs folder for your provider, and add it to the list in the main project readme.

## Vendoring Dependencies

If your provider depends on other go packages, then you must vendor them. To do this, use [govendor](https://github.com/kardianos/govendor).
