---
name: DNSKEY
parameters:
  - name
  - flags
  - protocol
  - algorithm
  - publicKey
  - modifiers...
parameter_types:
  name: string
  flags: number
  protocol: number
  algorithm: number
  publicKey: string
  "modifiers...": RecordModifier[]
---

DNSKEY adds a DNSKEY record to the domain.

Flags should be a number.

Protocol should be a number.

Algorithm must be a number.

Public key must be a string.

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  DNSKEY("@", 257, 3, 13, "AABBCCDD"),
END);
```
{% endcode %}
