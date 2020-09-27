---
name: AUTODNSSEC_ON
---

AUTODNSSEC_ON tells the provider to enable AutoDNSSEC.

AUTODNSSEC_OFF tells the provider to disable AutoDNSSEC.

AutoDNSSEC is a feature where a DNS provider can automatically manage
DNSSEC for a domain. Not all providers support this.

At this time, AUTODNSSEC_ON takes no parameters.  There is no ability
to tune what the DNS provider sets, no algorithm choice.  We simply
ask that they follow their defaults when enabling a no-fuss DNSSEC
data model.

NOTE: No parenthesis should follow these keywords.  That is, the
correct syntax is `AUTODNSSEC_ON` not `AUTODNSSEC_ON()`

{% include startExample.html %}
{% highlight js %}
D("example.com", .... ,
  AUTODNSSEC_ON,  // Enable AutoDNSSEC.
  A("@", "10.1.1.1")
);

D("insecure.com", .... ,
  AUTODNSSEC_OFF,  // Disable AutoDNSSEC.
  A("@", "10.2.2.2")
);
{%endhighlight%}
{% include endExample.html %}

If neither `AUTODNSSEC_ON` or `AUTODNSSEC_OFF` is specified for a
domain no changes will be requested.
