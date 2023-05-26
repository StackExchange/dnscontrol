---
name: INCLUDE
parameters:
  - domain
parameter_types:
  domain: string
---

Includes all records from a given domain


{% code title="dnsconfig.js" %}
```javascript
D("example.com!external", REG_MY_PROVIDER, DnsProvider(R53),
  A("test", "8.8.8.8")
);

D("example.com!internal", REG_MY_PROVIDER, DnsProvider(R53),
  INCLUDE("example.com!external"),
  A("home", "127.0.0.1")
);
```
{% endcode %}
