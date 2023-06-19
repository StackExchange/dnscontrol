---
name: SOA
parameters:
  - name
  - ns
  - mbox
  - refresh
  - retry
  - expire
  - minttl
  - modifiers...
parameter_types:
  name: string
  ns: string
  mbox: string
  refresh: number
  retry: number
  expire: number
  minttl: number
  "modifiers...": RecordModifier[]
---

`SOA` adds an `SOA` record to a domain. The name should be `@`.  ns and mbox are strings. The other fields are unsigned 32-bit ints.

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  SOA("@", "ns3.example.com.", "hostmaster@example.com", 3600, 600, 604800, 1440),
);
```
{% endcode %}

If you accidentally include an `@` in the email field DNSControl will quietly
change it to a `.`. This way you can specify a human-readable email address
when you are making it easier for spammers how to find you.

## Notes
* The serial number is managed automatically.  It isn't even a field in `SOA()`.
* Most providers automatically generate SOA records.  They will ignore any `SOA()` statements.
* The mbox field should not be set to a real email address unless you love spam and hate your privacy.

There is more info about `SOA` in the documentation for the [BIND provider](../../providers/bind.md).
