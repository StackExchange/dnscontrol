---
name: DOMAIN_ELSEWHERE
parameters:
  - registrar
  - nameserver_names
parameter_types:
  registrar: string
  nameserver_names: string[]
---

`DOMAIN_ELSEWHERE()` is a helper macro that lets you easily indicate that
a domain's zones are managed elsewhere. That is, it permits you easily delegate
a domain to a hard-coded list of DNS servers.

`DOMAIN_ELSEWHERE` is useful when you control a domain's registrar but not the
DNS servers. For example, suppose you own a domain but the DNS servers are run
by someone else, perhaps a SaaS product you've subscribed to or a DNS server
that is run by your brother-in-law who doesn't trust you with the API keys that
would let you maintain the domain using DNSControl. You need an easy way to
point (delegate) the domain at a specific list of DNS servers.

For example these two statements are equivalent:

```js
DOMAIN_ELSEWHERE("example.com", REG_NAMEDOTCOM, ["ns1.foo.com", "ns2.foo.com"]);

// ...is equivalent to...

D("example.com", REG_NAMEDOTCOM,
    NO_PURGE,
    NAMESERVER("ns1.foo.com"),
    NAMESERVER("ns2.foo.com")
);
```

NOTE: The `NO_PURGE` is used out of abundance of caution but since no
`DnsProvider()` statements exist, no updates would be performed.
