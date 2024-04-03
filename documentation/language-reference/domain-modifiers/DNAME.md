---
name: DNAME
parameters:
  - name
  - target
  - modifiers...
parameter_types:
  name: string
  target: string
  "modifiers...": RecordModifier[]
---

DNAME adds a DNAME record to the domain.

Target should be a string.

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  DNAME("sub", "example.net.")
);
```
{% endcode %}
