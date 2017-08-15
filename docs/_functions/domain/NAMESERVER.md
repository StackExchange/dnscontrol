---
name: NAMESERVER
parameters:
  - name
  - ip
  - modifiers...
---

NAMESERVER NS instructs DNSControl to inform the domain's registrar where to find this zone.
For some registrars this will also add NS records to the zone itself.

`ip` is optional, and is only required if glue records need to be generated in the parent zone.

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, .... ,
  NAMESERVER("ns1.myserver.com"),
  NAMESERVER("ns2.example.com", "100.100.100.100"), // the server plus glue
  A("www", "10.10.10.10"),
);

{%endhighlight%}
{% include endExample.html %}
