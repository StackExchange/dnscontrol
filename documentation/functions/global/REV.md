---
name: REV
parameters:
  - address
parameter_types:
  address: string
ts_return: string
---

`REV` returns the reverse lookup domain for an IP network. For
example `REV("1.2.3.0/24")` returns `3.2.1.in-addr.arpa.` and
`REV("2001:db8:302::/48")` returns `2.0.3.0.8.b.d.0.1.0.0.2.ip6.arpa.`.
This is used in [`D()`](D.md) functions to create reverse DNS lookup zones.

This is a convenience function. You could specify `D("3.2.1.in-addr.arpa",
...` if you like to do things manually but why would you risk making
typos?

`REV` complies with RFC2317, "Classless in-addr.arpa delegation"
for netmasks of size /25 through /31.
While the RFC permits any format, we abide by the recommended format:
`FIRST/MASK.C.B.A.in-addr.arpa` where `FIRST` is the first IP address
of the zone, `MASK` is the netmask of the zone (25-31 inclusive),
and A, B, C are the first 3 octets of the IP address. For example
`172.20.18.130/27` is located in a zone named
`128/27.18.20.172.in-addr.arpa`

If the address does not include a "/" then `REV` assumes /32 for IPv4 addresses
and /128 for IPv6 addresses.

Note that the lower bits (the ones outside the netmask) must be zeros. They are not
zeroed out automatically. Thus, `REV("1.2.3.4/24")` is an error.  This is done
to catch typos.

{% code title="dnsconfig.js" %}
```javascript
D(REV("1.2.3.0/24"), REGISTRAR, DnsProvider(BIND),
  PTR("1", "foo.example.com."),
  PTR("2", "bar.example.com."),
  PTR("3", "baz.example.com."),
  // These take advantage of DNSControl's ability to generate the right name:
  PTR("1.2.3.10", "ten.example.com."),
);

D(REV("2001:db8:302::/48"), REGISTRAR, DnsProvider(BIND),
  PTR("1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0", "foo.example.com."),  // 2001:db8:302::1
  // These take advantage of DNSControl's ability to generate the right name:
  PTR("2001:db8:302::2", "two.example.com."),                          // 2.0.0...
  PTR("2001:db8:302::3", "three.example.com."),                        // 3.0.0...
);
```
{% endcode %}

In the future we plan on adding a flag to [`A()`](../domain/A.md)which will insert
the correct PTR() record in the appropriate `D(REV())` domain (i.e. `.arpa` domain) has been
defined.
