---
name: CAA
parameters:
  - name
  - tag
  - value
  - modifiers...
---

CAA adds a CAA record to a domain. The name should be the relative label for the record. Use `@` for the domain apex.

Tag can be one of "issue", "issuewild" or "iodef".

Value is a string. The format of the contents is different depending on the tag.  DNSControl will handle any escaping or quoting required, similer to TXT records.  For example use `CAA("@", "issue", "letsencrypt.org")` rather than `CAA("@", "issue", "\"letsencrypt.org\"")`.

Flags are controlled by modifier.:

- CAA_CRITICAL: Issuer critical flag. CA that does not understand this tag will refuse to issue certificate for this domain.

CAA record is supported only by BIND and Google Cloud DNS. Some certificate authorities may not support this record until the mandatory date of September 2017.

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, DnsProvider("GCLOUD"),
  // Allow letsencrypt to issue certificate for this domain
  CAA("@", "issue", "letsencrypt.org"),
  // Allow no CA to issue wildcard certificate for this domain
  CAA("@", "issuewild", ";"),
  // Report all violation to test@example.com. If CA does not support
  // this record then refuse to issue any certificate
  CAA("@", "iodef", "mailto:test@example.com", CAA_CRITICAL)
);

{%endhighlight%}
{% include endExample.html %}
