---
name: ALIAS
parameters:
  - name
  - target
  - modifiers...
---

ALIAS is a virtual record type that points a record at another record. It is analagous to a CNAME, but is usually resolved at request-time and served as an A record. Unlike CNAMEs, ALIAS records can be used at the zone apex (`@`)

Different providers handle ALIAS records differently, and many do not support it at all. Attempting to use ALIAS records with a DNS provider type that does not support them will result in an error.

The name should be the relative label for the domain.

Target should be a string representing the target. If it is a single label we will assume it is a relative name on the current domain. If it contains *any* dots, it should be a fully qualified domain name, ending with a `.`.

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, DnsProvider("CLOUDFLARE"),
  ALIAS("@", "google.com."), // example.com -> google.com
);

{%endhighlight%}
{% include endExample.html %}