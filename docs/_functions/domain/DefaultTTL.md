---
name: DefaultTTL
parameters:
  - ttl
---

DefaultTTL sets the TTL for all records in a domain that do not explicitly set one with [TTL](#TTL). If neither `DefaultTTl` or `TTL` exist for a record,
it will use the DNSControl global default of 300 seconds.

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, DnsProvider("R53"),
  DefaultTTL(2000),
  A("@","1.2.3.4"), //has default
  A("foo", "2.3.4.5", TTL(500)) //overrides default
);

{%endhighlight%}
{% include endExample.html %}