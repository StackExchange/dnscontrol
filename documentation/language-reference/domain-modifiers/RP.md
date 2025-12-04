---
name: RP
parameters:
  - name
  - mbox
  - txt
  - modifiers...
parameter_types:
  name: string
  mbox: string
  txt: string
  "modifiers...": RecordModifier[]
---

`RP()` adds an RP record to a domain.

The RP implementation in DNSControl is still experimental and may change.

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  RP("@", "user.example.com.", "example.com."),
);
```
{% endcode %}
