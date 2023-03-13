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
D("example.com", REG_THIRDPARTY, DnsProvider("DNS_BIND"),
  SOA("@", "ns3.example.org.", "hostmaster@example.org", 3600, 600, 604800, 1440),
);
```
{% endcode %}

The email address should be specified like a normal RFC822/RFC5322 address (user@hostname.com). It will be converted into the required format (e.g. BIND format: `user.hostname.com`) by the provider as required. This has the benefit of being more human-readable plus DNSControl can properly handle escaping and other issues.

## Notes
* Previously, the accepted format for the SOA mailbox field was `hostmaster.example.org`. This has been changed to `hostmaster@example.org`
* The serial number is managed automatically.  It isn't even a field in `SOA()`.
* Most providers automatically generate SOA records.  They will ignore any `SOA()` statements.

There is more info about SOA in the documentation for the [BIND provider](../../providers/bind.md).
