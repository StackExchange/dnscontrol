---
name: TLSA
parameters:
  - name
  - usage
  - selector
  - type
  - certificate
  - modifiers...
---

TLSA adds a TLSA record to a domain. The name should be the relative label for the record.

Usage, selector, and type are ints.

Certificate is a hex string.

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, DnsProvider("GCLOUD"),
  // Create TLSA record for certificate used on TCP port 443
  TLSA("_443._tcp", 3, 1, 1, "abcdef0"),
);

{%endhighlight%}
{% include endExample.html %}
