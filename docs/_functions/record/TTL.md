---
name: TTL
parameters:
  - ttl
---

TTL sets the TTL for a single record only. This will take precedence
over the domain's [DefaultTTL](#DefaultTTL) if supplied.

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, DnsProvider("R53"),
  DefaultTTL(2000),
  A("@","1.2.3.4"), //has default
  A("foo", "2.3.4.5", TTL(500)) //overrides default
);

{%endhighlight%}
{% include endExample.html %}