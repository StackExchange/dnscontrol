---
name: NAMESERVER_TTL
parameters:
  - ttl
---

TTL sets the TTL on the domain apex NS RRs defined by [NAMESERVER](#NAMESERVER).

The value can be an integer or a string. See [TTL](#TTL) for examples.

{% include startExample.html %}
{% highlight js %}

D('example.com', REGISTRAR, DnsProvider('R53'),
  NAMESERVER_TTL('2d'),
  NAMESERVER('ns')
);
{%endhighlight%}
{% include endExample.html %}
