---
name: IGNORE_TARGET
parameters:
  - pattern
  - rType
---

IGNORE_TARGET can be used to ignore some records present in zone based on the record's target and type. IGNORE_TARGET currently only supports CNAME record types.

IGNORE_TARGET is like NO_PURGE except it acts only on some specific records instead of the whole zone.

IGNORE_TARGET is generally used in very specific situations:

* Some records are managed by some other system and DNSControl is only used to manage some records and/or keep them updated. For example a DNS record that is created by AWS Certificate Manager for validation, but DNSControl is used to manage the rest of the zone. In this case we don't want DNSControl to try to delete the externally managed record.

In this example, DNSControl will insert/update the "baz.example.com" record but will leave unchanged a CNAME to "foo.acm-validations.aws" record.

{% include startExample.html %}
{% highlight js %}
D("example.com",
  IGNORE_TARGET('**.acm-validations.aws.', 'CNAME'),
  A("baz", "1.2.3.4")
);
{%endhighlight%}
{% include endExample.html %}

IGNORE_TARGET also supports glob patterns in the style of the [gobwas/glob](https://github.com/gobwas/glob#example) library. Some example patterns:

* `IGNORE_TARGET("example.com", "CNAME")` will ignore all CNAME records with targets of exactly `example.com`.
* `IGNORE_TARGET("*.foo", "CNAME")` will ignore all CNAME records with targets in the style of `bar.foo`, but will not ignore records with targets using a double subdomain, such as `foo.bar.foo`.
* `IGNORE_TARGET("**.bar", "CNAME")` will ignore all CNAME records with target subdomains of `bar`, including double subdomains such as `www.foo.bar`.
* `IGNORE_TARGET("dev.*.foo", "CNAME")` will ignore all CNAME records with targets in the style of `dev.bar.foo`, but will not ignore records with targets using a double subdomain, such as `dev.foo.bar.foo`.

It is considered as an error to try to manage an ignored record.