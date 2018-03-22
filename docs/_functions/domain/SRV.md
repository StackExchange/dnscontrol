---
name: SRV
parameters:
  - name
  - priority
  - weight
  - port
  - target
  - modifiers...
---

`SRV` adds a `SRV` record to a domain. The name should be the relative label for the record.

Priority, weight, and port are ints.

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, DnsProvider("GCLOUD"),
  // Create SRV records for a a SIP service:
  //               pr  w   port, target
  SRV('_sip._tcp', 10, 60, 5060, 'bigbox.example.tld.'),
  SRV('_sip._tcp', 10, 20, 5060, 'smallbox1.example.tld.'),
);

{%endhighlight%}
{% include endExample.html %}
