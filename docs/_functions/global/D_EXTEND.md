---
name: D_EXTEND
parameters:
  - name
  - modifiers...
---

`D_EXTEND` adds records (and metadata) to a domain. The parent domain must have previously been defined by `D()`. As with `D()`, the first argument to `D_EXTEND()` is the domain name. The domain name provided to `D_EXTEND()` may also include non-delegated subdomain parts. If a subdomain is provided the subdomain part will be appended to all record names, with the exception of `CF_REDIRECT` and `CF_TEMP_REDIRECT` which are always on the apex domain. See the documentation of `D` for further details.

Example:

{% include startExample.html %}
{% highlight js %}
D('domain.tld', REG, DnsProvider(DNS),
  A('@', "127.0.0.1")         // domain.tld
  A('www', "127.0.0.2")       // www.domain.tld
)
D_EXTEND('domain.tld',
  A('aaa', "127.0.0.3")       // aaa.domain.tld
)
D_EXTEND('sub.domain.tld',
  A('bbb', "127.0.0.4"),      // bbb.sub.domain.tld
  A('ccc', "127.0.0.5")       // ccc.sub.domain.tld
)
D_EXTEND('sub.sub.domain.tld',
  A('ddd', "127.0.0.6")       // ddd.sub.sub.domain.tld
)
{%endhighlight%}

This will end up in the following modifications:
```
******************** Domain: domain.tld
----- Getting nameservers from: registrar
----- DNS Provider: registrar...3 corrections
#1: CREATE A domain.tld 127.0.0.1 ttl=43200
#2: CREATE A www.domain.tld 127.0.0.2 ttl=43200
#3: CREATE A aaa.domain.tld 127.0.0.3 ttl=43200
#4: CREATE A bbb.sub.domain.tld 127.0.0.4 ttl=43200
#5: CREATE A ccc.sub.domain.tld 127.0.0.5 ttl=43200
#5: CREATE A ddd.sub.sub.domain.tld 127.0.0.6 ttl=43200
#6: REFRESH zone domain.tld
```
{% include endExample.html %}

ProTips: `D_EXTEND()` permits you to create very complex and
sophisticated configurations, but you shouldn't. Be nice to the next
person that edits the file, who may not be as expert as yourself.
Enhance readability by putting any `D_EXTEND()` statements immediately
after the original `D()`, like in above example.  Avoid the temptation
to obscure the addition of records to existing domains with randomly
placed `D_EXTEND()` statements. Don't build up a domain using loops of
`D_EXTEND()` statements. You'll be glad you didn't.
