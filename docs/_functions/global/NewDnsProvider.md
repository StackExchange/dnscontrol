---
name: NewDnsProvider
parameters:
  - name
  - type
  - meta
return: string
---

NewDnsProvider activates a DNS Service Provider (DSP) specified in creds.json.
A DSP stores a DNS zone's records and provides DNS service for the zone (i.e.
answers on port 53 to queries related to the zone).

* `name` must match the name of an entry in `creds.json`.
* `type` specifies a valid DNS provider type identifier listed on the [provider page]({{site.github.url}}/provider-list).
  * Starting with v3.16, the type is optional. If it is absent, the `TYPE` field in `creds.json` is used instead. You can leave it out. (Thanks to JavaScript magic, you can leave it out even when there are more fields).
  * Starting with v4.0, specifying the type may be an error. Please add the `TYPE` field to `creds.json` and remove this parameter from `dnsconfig.js` to prepare.
* `meta` is a way to send additional parameters to the provider.  It is optional and only certain providers use it.  See the [individual provider docs]({{site.github.url}}/provider-list) for details.

This function will return an opaque string that should be assigned to a variable name for use in [D](#D) directives.

Prior to v3.16:

```js
var REG_MYNDC = NewRegistrar("mynamedotcom", "NAMEDOTCOM");
var DNS_MYAWS = NewDnsProvider("myaws", "ROUTE53");

D("example.com", REG_MYNDC, DnsProvider(DNS_MYAWS),
  A("@","1.2.3.4")
);
```

In v3.16 and later:

```js
var REG_MYNDC = NewRegistrar("mynamedotcom");
var DNS_MYAWS = NewDnsProvider("myaws");

D("example.com", REG_MYNDC, DnsProvider(DNS_MYAWS),
  A("@","1.2.3.4")
);
```
