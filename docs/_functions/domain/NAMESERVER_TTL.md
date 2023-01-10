---
name: NAMESERVER_TTL
parameters:
  - ttl
---

TTL sets the TTL on the domain apex NS RRs defined by [NAMESERVER](#NAMESERVER).

The value can be an integer or a string. See [TTL](#TTL) for examples.

{% capture example %}
```js
D('example.com', REGISTRAR, DnsProvider('R53'),
  NAMESERVER_TTL('2d'),
  NAMESERVER('ns')
);
```
{% endcapture %}

{% include example.html content=example %}
