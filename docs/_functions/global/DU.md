---
name: DU
parameters:
  - name
  - modifiers...
---

`DU` allows updating existing domains added by `D`. It behaves the same way, however it is mandatory that the domain was already added. The first argument, the domain name,
is required. Check out the documentation of `D` for further details.

Example:
{% include startExample.html %}
{% highlight js %}
D('domain.tld', REG, DnsProvider(DNS),
  A('@', "127.0.0.1")
)
DU('domain.tld',
  A('@', "127.0.0.2")
)
{%endhighlight%}

This will end up in following modifications:
```
******************** Domain: domain.tld
----- Getting nameservers from: registrar
----- DNS Provider: registrar...3 corrections
#1: CREATE A domain.tld 127.0.0.1 ttl=43200
#2: CREATE A domain.tld 127.0.0.2 ttl=43200
#3: REFRESH zone domain.tld
```
{% include endExample.html %}
