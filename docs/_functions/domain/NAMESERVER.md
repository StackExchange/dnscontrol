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


# The difference between NS() and NAMESERVER()

Nameservers are one of the least
understood parts of DNS, so a little extra explanation is required.

* `NS()` lets you add an NS record to a zone, just like A() adds an A
  record to the zone.

* The `NAMESERVER()` directive adds an NS record to the parent zone.

Since the parent zone could be completely unrelated to the current
zone, changes made by `NAMESERVER()` have to be done by an API call to
the registrar, who then figures out what to do. For example, if I
change the `NAMESERVER()` for stackoverflow.com, DNSControl talks to
the registrar who does the hard work of talking to the people that
control `.com`.  If the domain was gmeet.io, the registrar does
the right thing to talk to the people that control `.io`.

(Maybe it should have been called `PARENTNAMESERVER()` but we didn't
think of that at the time.)

When you use `NAMESERVER()`, DNSControl takes care of adding the
appropriate `NS` records to the zone.

Therefore, you don't have to specify `NS()` records except when
delegating a subdomain, in which case you are acting like a registrar!

Many DNS Providers will handle all of this for you, pick the name of
the nameservers for you and updating them (upward and in your zone)
automatically.  For more information, refer to
[this page]({{site.github.url}}/nameservers).


That's why NAMESERVER() is a separate operator.


# How to not change the parent NS records?

If dnsconfig.js has zero `NAMESERVER()` commands for a domain, it will
use the API to remove all the nameservers.

If dnsconfig.js has 1 or more `NAMESERVER()` commands for a domain, it
will use the API to set those as the nameservers (unless, of course,
they're already correct).

So how do you tell DNSControl not to make any changes?  Use the
special Registrar called "NONE". It makes no changes.

It looks like this:

```
var REG_THIRDPARTY = NewRegistrar('ThirdParty', 'NONE')
D("mydomain.com", REG_THIRDPARTY,
  ...
)
```
