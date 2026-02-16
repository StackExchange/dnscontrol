---
name: MIKROTIK_NXDOMAIN
parameters:
  - name
  - modifiers...
parameter_types:
  name: string
  "modifiers...": RecordModifier[]
provider: MIKROTIK
---

`MIKROTIK_NXDOMAIN` creates a RouterOS NXDOMAIN static entry. The router will respond with NXDOMAIN for any DNS queries matching the specified name. This is commonly used for DNS-based ad blocking or blackholing.

Metadata keys supported:

| Key               | Description                                                        |
|-------------------|--------------------------------------------------------------------|
| `match_subdomain` | Set to `"true"` to also match subdomains of the name.              |
| `regexp`          | RouterOS regexp for query matching.                                |
| `comment`         | Comment stored on the RouterOS record.                             |

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
    MIKROTIK_NXDOMAIN("ads"),
);
```
{% endcode %}
