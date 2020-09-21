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
in the current zone and accepts a label. `NS()` is useful for downward
delegations. `NAMESERVER()` is for informing upstream delegations.

For more information, refer to [this page]({{site.github.url}}/nameservers).

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, .... ,
  DnsProvider(route53, 0),
  // Replace the nameservers:
  NAMESERVER("ns1.myserver.com."),
  NAMESERVER("ns2.myserver.com."),
);

D("example2.com", REGISTRAR, .... ,
  // Add these two additional nameservers to the existing list of nameservers.
  NAMESERVER("ns1.myserver.com."),
  NAMESERVER("ns2.myserver.com."),
);

{%endhighlight%}
{% include endExample.html %}


# The difference between NS() and NAMESERVER()

Nameservers are one of the least
understood parts of DNS, so a little extra explanation is required.

* `NS()` lets you add an NS record to a zone, just like A() adds an A
  record to the zone. This is generally used to delegate a subzone.

* The `NAMESERVER()` directive speaks to the Registrar about how the parent should delegate the zone.

Since the parent zone could be completely unrelated to the current
zone, changes made by `NAMESERVER()` have to be done by an API call to
the registrar, who then figures out what to do. For example, if I
use `NAMESERVER()` in the zone `stackoverflow.com`, DNSControl talks to
the registrar who does the hard work of talking to the people that
control `.com`.  If the domain was `gmeet.io`, the registrar does
the right thing to talk to the people that control `.io`.

(A better name might have been `PARENTNAMESERVER()` but we didn't
think of that at the time.)

Each registrar handles delegations differently.  Most use
the `NAMESERVER()` targets to update the delegation, adding
`NS` records to the parent zone as required.
Some providers restrict the names to hosts they control.
Others may require you to add the `NS` records to the parent domain
manually.

# How to not change the parent NS records?

If dnsconfig.js has zero `NAMESERVER()` commands for a domain, it will
use the API to remove all non-default nameservers.

If dnsconfig.js has 1 or more `NAMESERVER()` commands for a domain, it
will use the API to add those nameservers (unless, of course,
they already exist).

So how do you tell DNSControl not to make any changes at all?  Use the
special Registrar called "NONE". It makes no changes.

It looks like this:

```
var REG_THIRDPARTY = NewRegistrar('ThirdParty', 'NONE')
D("mydomain.com", REG_THIRDPARTY,
  ...
)
```
