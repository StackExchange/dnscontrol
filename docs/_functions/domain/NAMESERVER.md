---
name: NAMESERVER
parameters:
  - name
  - modifiers...
---

`NAMESERVER()` instructs DNSControl to inform the domain's registrar where to find this zone.
For some registrars this will also add NS records to the zone itself.

This takes exactly one argument: the name of the nameserver. It must end with
a "." if it is a FQDN, just like all targets.

This is different than the `NS()` function, which inserts NS records
in the current zone and accepts a label. It is useful for downward
delegations. This is for informing upstream delegations.

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, .... ,
  NAMESERVER("ns1.myserver.com."),
  NAMESERVER("ns2.myserver.com."),
);

{%endhighlight%}
{% include endExample.html %}
