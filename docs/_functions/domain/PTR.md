---
name: PTR
parameters:
  - name
  - target
  - modifiers...
---

PTR adds a PTR record to the domain.

The name should be the relative label for the domain, or may be a FQDN that ends with `.`.

* If the name is a valid IP address, DNSControl will *magically* replace it with a string that is appropriate for the domain. That is, if the domain ends with `in-addr.arpa` it will generate the IPv4-style reverse name; if the domain ends with `ipv6.arpa` it will generate the IPv6-style reverse name.  DNSControl will truncate it as appropriate for the netmask.
* If the name ends with `in-addr.arpa.` or `ipv6.arpa.` (not the `.` at the end), DNSControl will truncate it as appropriate for the domain. If the FQDN does not fit within the domain, this will raise an error.

Target should be a string representing the FQDN of a host.  Like all FQDNs in DNSControl, it must end with a `.`.

{% include startExample.html %}
{% highlight js %}
D(REV('1.2.3.0/24'), REGISTRAR, DnsProvider(BIND),
  PTR('1', 'foo.example.com.'),
  PTR('2', 'bar.example.com.'),
  PTR('3', 'baz.example.com.'),
  // If the first parameter is a valid IP address, DNSControl will generate the correct name:
  PTR('1.2.3.10', 'ten.example.com.'),    // '10'
);

D(REV('2001:db8:302::/48'), REGISTRAR, DnsProvider(BIND),
  PTR('1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0', 'foo.example.com.'),  // 2001:db8:302::1
  // If the first parameter is a valid IP address, DNSControl will generate the correct name:
  PTR('2001:db8:302::2', 'two.example.com.'),                          // '2.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0'
  PTR('2001:db8:302::3', 'three.example.com.'),                        // '3.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0'
);

{%endhighlight%}
{% include endExample.html %}

In the future we plan on adding a flag to `A()` which will insert
the correct PTR() record if the approprate `.arpa` domain has been
defined.
