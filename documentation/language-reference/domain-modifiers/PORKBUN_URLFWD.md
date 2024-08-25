---
name: PORKBUN_URLFWD
parameters:
  - name
  - target
  - modifiers...
provider: PORKBUN
parameter_types:
  name: string
  target: string
  "modifiers...": RecordModifier[]
---

`PORKBUN_URLFWD` is a Porkbun-specific feature that maps to Porkbun's URL forwarding feature, which creates HTTP 301 (permanent) or 302 (temporary) redirects.


{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
    PORKBUN_URLFWD("urlfwd1", "http://example.com"),
    PORKBUN_URLFWD("urlfwd2", "http://example.org", {type: "permanent", includePath: "yes", wildcard: "no"})
);
```
{% endcode %}

The fields are:
* name: the record name
* target: where you'd like to forward the domain to
* type: valid types are: `temporary` (302 / 307) or `permanent` (301), default to `temporary`
* includePath: whether to include the URI path in the redirection. Valid options are `yes` or `no`, default to `no`
* wildcard: forward all subdomains of the domain. Valid options are `yes` or `no`, default to `yes`
