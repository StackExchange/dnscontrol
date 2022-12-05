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

{% capture example %}
```js
D("example.com", REG_THIRDPARTY, DnsProvider("DNS_BIND"),
  SOA("@", "ns3.example.org.", "hostmaster.example.org.", 3600, 600, 604800, 1440),
);
```
{% endcapture %}

{% include example.html content=example %}

## Notes:

* The serial number is managed automatically.  It isn't even a field in `SOA()`.
* Most providers automatically generate SOA records.  They will ignore any `SOA()` statements.

There is more info about SOA in the documentation for the [BIND provider]({{site.github.url}}/providers/bind).
