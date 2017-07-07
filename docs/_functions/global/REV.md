---
name: REV
parameters:
  - address
---

`REV` returns the reverse lookup domain for an IP network. For example `REV('1.2.3.0/24')` returns `3.2.1.in-addr.arpa.`
and `REV('2001:db8:302::/48)` returns `2.0.3.0.8.b.d.0.1.0.0.2.ip6.arpa.`. This is used in `D()` functions to create
reverse DNS (`PTR`) zones.

This is a convenience function. You could specify `D('3.2.1.in-addr.arpa`, ...` if you like to do things manually
and permit typos to creep in.

The network portion of the IP address (`/24`) must always be specified.

Note that the lower bits are zeroed out automatically. Thus, `REV('1.2.3.4/24') is the same as `REV('1.2.3.0/24')`. This
may generate warnings or errors in the future.

{% include startExample.html %}
{% highlight js %}
D(REV('1.2.3.0/24'), REGISTRAR, DnsProvider(BIND),
  PTR("1", 'foo.example.com.'),
  PTR("2", 'bar.example.com.'),
  PTR("3", 'baz.example.com.'),
  // These take advantage of DNSControl's ability to generate the right name:
  PTR("1.2.3.10", 'ten.example.com.'),
);

D(REV('2001:db8:302::/48'), REGISTRAR, DnsProvider(BIND),
  PTR("1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0", 'foo.example.com.'),  // 2001:db8:302::1
  // These take advantage of DNSControl's ability to generate the right name:
  PTR("2001:db8:302::2", 'two.example.com.'),                          // 2.0.0. etc. etc.
  PTR("2001:db8:302::3", 'three.example.com.'),                        // 
);


{%endhighlight%}
{% include endExample.html %}

In the future we plan on adding a flag to `A()` which will insert 
the correct PTR() record if the approprate `D(REV()` domain (i.e. `.arpa` domain) has been
defined.
