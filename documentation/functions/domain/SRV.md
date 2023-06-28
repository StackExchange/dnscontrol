---
name: SRV
parameters:
  - name
  - priority
  - weight
  - port
  - target
  - modifiers...
parameter_types:
  name: string
  priority: number
  weight: number
  port: number
  target: string
  "modifiers...": RecordModifier[]
---

`SRV` adds a `SRV` record to a domain. The name should be the relative label for the record.

Priority, weight, and port are ints.

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  // Create SRV records for a a SIP service:
  //               pr  w   port, target
  SRV("_sip._tcp", 10, 60, 5060, "bigbox.example.com."),
  SRV("_sip._tcp", 10, 20, 5060, "smallbox1.example.com."),
);
```
{% endcode %}
