---
name: D_EXTEND
parameters:
  - name
  - modifiers...
---

`D_EXTEND` adds records (and metadata) to a domain. The parent domain must have previously been defined by `D()`. `D_EXTEND()` behaves the same as `D()` in all other ways: The first argument is the domain name, which may include non-delegated subdomain parts. If a subdomain is provided the subdomain part will be appended to all record names. See the documentation of `D` for further details.

Example:
{% include startExample.html %}
{% highlight js %}
D('domain.tld', REG, DnsProvider(DNS),
  A('@', "127.0.0.1")
)
D_EXTEND('domain.tld',
  A('@', "127.0.0.2")
)
D_EXTEND('sub.domain.tld',
  A('@', "127.0.0.3"),
  A('a', "127.0.0.4")
)
D_EXTEND('sub.sub.domain.tld',
  A('a', "127.0.0.5")
)
{%endhighlight%}

This will end up in following modifications:
```
******************** Domain: domain.tld
----- Getting nameservers from: registrar
----- DNS Provider: registrar...3 corrections
#1: CREATE A domain.tld 127.0.0.1 ttl=43200
#2: CREATE A domain.tld 127.0.0.2 ttl=43200
#3: CREATE A sub.domain.tld 127.0.0.3 ttl=43200
#4: CREATE A a.sub.domain.tld 127.0.0.4 ttl=43200
#5: CREATE A a.sub.sub.domain.tld 127.0.0.5 ttl=43200
#6: REFRESH zone domain.tld
```
{% include endExample.html %}
