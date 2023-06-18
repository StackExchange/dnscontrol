---
name: AAAA
parameters:
  - name
  - address
  - modifiers...
parameter_types:
  name: string
  address: string
  "modifiers...": RecordModifier[]
---

AAAA adds an AAAA record To a domain. The name should be the relative label for the record. Use `@` for the domain apex.

The address should be an IPv6 address as a string.

Modifiers can be any number of [record modifiers](https://docs.dnscontrol.org/language-reference/record-modifiers) or JSON objects, which will be merged into the record's metadata.

{% code title="dnsconfig.js" %}
```javascript
var addrV6 = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"

D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  AAAA("@", addrV6),
  AAAA("foo", addrV6),
  AAAA("test.foo", addrV6, TTL(5000)),
  AAAA("*", addrV6, {foo: 42})
);
```
{% endcode %}
