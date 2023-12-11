---
name: DHCID
parameters:
  - name
  - digest
  - modifiers...
parameter_types:
  name: string
  digest: string
  "modifiers...": RecordModifier[]
---

DHCID adds a DHCID record to the domain.

Digest should be a string.

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  DHCID("example.com", "ABCDEFG")
);
```
{% endcode %}
