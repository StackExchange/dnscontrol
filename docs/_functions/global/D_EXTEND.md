---
name: D_EXTEND
parameters:
  - name
  - modifiers...
---

`D_EXTEND` adds records (and metadata) to a domain previously defined
by `D()`, optionally adding subdomain records.

The first argument is a domain name. If it exactly matches a
previously defined domain, `D_EXTEND()` behaves the same as `D()`,
simply adding records as if they had been specified in the original
`D()`.

If the domain name does not match an existing domain, but could be a
(non-delegated) subdomain of an existing domain, the new records (and
metadata) are added with the subdomain part appended to all record
names. See the examples below.

Matching the domain name to previously-defined domains is done using a
`longest match` algorithm.  If `domain.tld` and `sub.domain.tld` are
defined as separate domains via separate `D()` statements, then
`D_EXTEND('sub.sub.domain.tld', ...)` would match `sub.domain.tld`,
not `domain.tld`.

`CF_REDIRECT` and `CF_TEMP_REDIRECT` are always on the apex domain, not
any subdomain.

Example:

{% include startExample.html %}
{% highlight js %}
D('domain.tld', REG, DnsProvider(DNS),
  A('@', "127.0.0.1"),        // domain.tld
  A('www', "127.0.0.2")       // www.domain.tld
);
D_EXTEND('domain.tld',
  A('aaa', "127.0.0.3")       // aaa.domain.tld
);
D_EXTEND('sub.domain.tld',
  A('bbb', "127.0.0.4"),      // bbb.sub.domain.tld
  A('ccc', "127.0.0.5")       // ccc.sub.domain.tld
);
D_EXTEND('sub.sub.domain.tld',
  A('ddd', "127.0.0.6")       // ddd.sub.sub.domain.tld
);
D_EXTEND('sub.domain.tld',
  A('@', "127.0.0.7")         // sub.domain.tld
);
{%endhighlight%}

This will end up in the following modifications:
```
******************** Domain: domain.tld
----- Getting nameservers from: cloudflare
----- DNS Provider: cloudflare...7 corrections
#1: CREATE record: aaa A 1 127.0.0.3
#2: CREATE record: bbb.sub A 1 127.0.0.4
#3: CREATE record: ccc.sub A 1 127.0.0.5
#4: CREATE record: ddd.sub.sub A 1 127.0.0.6
#5: CREATE record: @ A 1 127.0.0.1
#6: CREATE record: sub A 1 127.0.0.7
#7: CREATE record: www A 1 127.0.0.2
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
