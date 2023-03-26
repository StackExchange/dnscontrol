---
name: NewRegistrar
parameters:
  - name
  - type
  - meta
parameter_types:
  name: string
  type: string?
  meta: object?
return: string
---

NewRegistrar activates a Registrar Provider specified in `creds.json`.
A registrar maintains the domain's registration and delegation (i.e. the
nameservers for the domain).  DNSControl only manages the delegation.

* `name` must match the name of an entry in `creds.json`.
* `type` specifies a valid DNS provider type identifier listed on the [provider page](../../providers.md).
  * Starting with [v3.16](../../v316.md), the type is optional. If it is absent, the `TYPE` field in `creds.json` is used instead. You can leave it out. (Thanks to JavaScript magic, you can leave it out even when there are more fields).
  * Starting with v4.0, specifying the type may be an error. Please add the `TYPE` field to `creds.json` and remove this parameter from `dnsconfig.js` to prepare.
* `meta` is a way to send additional parameters to the provider.  It is optional and only certain providers use it.  See the [individual provider docs](../../providers.md) for details.

This function will return an opaque string that should be assigned to a variable name for use in [D](D.md) directives.

Prior to [v3.16](../../v316.md):

{% code title="dnsconfig.js" %}
```javascript
var REG_MYNDC = NewRegistrar("mynamedotcom", "NAMEDOTCOM");
var DNS_MYAWS = NewDnsProvider("myaws", "ROUTE53");

D("example.com", REG_MYNDC, DnsProvider(DNS_MYAWS),
  A("@","1.2.3.4")
);
```
{% endcode %}

In [v3.16](../../v316.md) and later:

{% code title="dnsconfig.js" %}
```javascript
var REG_MYNDC = NewRegistrar("mynamedotcom");
var DNS_MYAWS = NewDnsProvider("myaws");

D("example.com", REG_MYNDC, DnsProvider(DNS_MYAWS),
  A("@","1.2.3.4")
);
```
{% endcode %}
