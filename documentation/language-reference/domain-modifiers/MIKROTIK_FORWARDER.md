---
name: MIKROTIK_FORWARDER
parameters:
  - name
  - dns_servers
  - modifiers...
parameter_types:
  name: string
  dns_servers: string
  "modifiers...": RecordModifier[]
provider: MIKROTIK
---

`MIKROTIK_FORWARDER` manages a RouterOS DNS forwarder entry (`/ip/dns/forwarders`). The `name` parameter can be a domain name (e.g. `corp.example.com`) or an arbitrary alias (e.g. `my-upstream`). These named entries can then be referenced as the target of [`MIKROTIK_FWD`](MIKROTIK_FWD.md) records.

Forwarder records must be placed in the synthetic zone `_forwarders.mikrotik`. This zone should appear **before** any zones that reference its entries by name in `dnsconfig.js` to ensure proper creation order.

Metadata keys supported:

| Key                | Description                                        |
|--------------------|----------------------------------------------------|

| `doh_servers`      | DoH server URLs for this forwarder.                |
| `verify_doh_cert`  | Set to `"true"` to verify the DoH certificate.     |

{% code title="dnsconfig.js" %}
```javascript
D("_forwarders.mikrotik", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
    MIKROTIK_FORWARDER("corp.example.com", "10.0.0.53,10.0.0.54"),
    MIKROTIK_FORWARDER("my-upstream", "1.1.1.1"),
);

// Then reference the alias in a FWD record:
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
    MIKROTIK_FWD("@", "my-upstream", {match_subdomain: "true"}),
);
```
{% endcode %}
