---
name: PANIC
parameters:
  - message
parameter_types:
  message: string
ts_return: never
---

`PANIC` terminates the script and therefore DNSControl with an exit code of 1. This should be used if your script cannot gather enough information to generate records, for example when a HTTP request failed.

{% code title="dnsconfig.js" %}
```javascript
PANIC("Something really bad has happened");
```
{% endcode %}
