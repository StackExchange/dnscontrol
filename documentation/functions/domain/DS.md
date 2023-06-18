---
name: DS
parameters:
  - name
  - keytag
  - algorithm
  - digesttype
  - digest
  - modifiers...
parameter_types:
  name: string
  keytag: number
  algorithm: number
  digesttype: number
  digest: string
  "modifiers...": RecordModifier[]
---

DS adds a DS record to the domain.

Key Tag should be a number.

Algorithm should be a number.

Digest Type must be a number.

Digest must be a string.

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  DS("example.com", 2371, 13, 2, "ABCDEF")
);
```
{% endcode %}
