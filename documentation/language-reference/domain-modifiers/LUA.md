---
name: LUA
parameters:
- name
- rtype
- contents
- modifiers...
parameter_types:
  name: string
  rtype: string
  contents: string | string[]
  "modifiers...": RecordModifier[]
---


# LUA

`LUA()` adds a **PowerDNS Lua record** to a domain. Use this when you want answers computed at **query time** (traffic steering, geo/ASN steering, weighted pools, health-based failover, time-based values, etc.) using the PowerDNS Authoritative Server’s built-in Lua engine.

> **Provider support:** `LUA()` is supported **only** by the **PowerDNS** DNS provider in DNSControl. Ensure your zones are served by PowerDNS and that Lua records are enabled.
> See: PowerDNS provider page and Supported providers matrix.
> (References at the end.)

## Signature

{% code title="dnsconfig.js" %}
```typescript
LUA(
  name: string,
  rtype: string,                  // e.g. "A", "AAAA", "CNAME", "TXT", "PTR", "LOC", ...
  contents: string | string[],    // the Lua snippet
  ...modifiers: RecordModifier[]
): DomainModifier
```
{% endcode %}

- **`name`** — label for the record (`"@"` for the zone apex).
- **`rtype`** — the RR type the Lua snippet **emits** (e.g., `"A"`, `"AAAA"`, `"CNAME"`, `"TXT"`, `"PTR"`, `"LOC"`).
- **`contents`** — the Lua snippet (string or array). See **Syntax** below.
- **`modifiers`** — standard record modifiers like `TTL(60)`.

## Prerequisites (PowerDNS)

PowerDNS Authoritative Server **4.2+** supports Lua records. You must enable Lua records either **globally** (in `pdns.conf`) or **per-zone** via domain metadata.

- **Global:** set `enable-lua-records=yes` (or `shared`) and reload PowerDNS.
- **Per-zone:** set metadata `ENABLE-LUA-RECORDS = 1` for the zone.

See PowerDNS’s **Lua Records** overview and **Lua Reference** for details and helpers.

## Syntax

PowerDNS evaluates the `contents` with two modes:

- **Single expression (most common):** write **just the expression** — **no `return`**. PowerDNS implicitly treats the snippet as if it were the argument to `return`.
- **Multi-statement script:** start the snippet with a **leading semicolon (`;`)**. In this mode you can write multiple statements and must include your own `return`.

The value produced must be valid **RDATA** for the chosen `rtype` (IPv4 for `A`, IPv6 for `AAAA`, a single FQDN with trailing dot for `CNAME`, proper text for `TXT`, etc.). Helper functions and preset variables (e.g., `pickrandom`, `pickclosest`, `country`, `continent`, `qname`, `ifportup`) are defined in the PowerDNS Lua reference.

## Examples

### Single expression (implicit `return`)

{% code title="dnsconfig.js" %}
```javascript
// Weighted/random selection
LUA("app", "A", "pickrandom({'192.0.2.11',3}, {'192.0.2.22',1})", TTL(60));

// Health-aware pool: only addresses with TCP/443 up are served
LUA("www", "A", "ifportup(443, {'192.0.2.1','192.0.2.2'})", TTL(60));

// Geo proximity
LUA("edge", "A", "pickclosest({'192.0.2.1','192.0.2.2','198.51.100.1'})", TTL(60));
```
{% endcode %}

### Multi-statement (leading `;`)

{% code title="dnsconfig.js" %}
```javascript
LUA("api", "A", `
  ; if continent('EU') then
      return {'198.51.100.1'}
    else
      return {'192.0.2.10','192.0.2.20'}
    end
`, TTL(60));

// Dynamic TXT, showing the queried name (string building example)
LUA("_diag", "TXT", "; return 'Got a TXT query for ' .. qname:toString()", TTL(30));
```
{% endcode %}

### Other RR types

Lua can emit data for many RR types as long as the RDATA is valid for that type:

{% code title="dnsconfig.js" %}
```javascript
LUA("edge", "CNAME", "('edgesvc.example.net.')", TTL(60));
LUA("pop.asu", "LOC", "latlonloc(-25.286, -57.645, 100)", TTL(300)); // ~Asunción, 100m
```
{% endcode %}

## Tips & gotchas

- **Use low TTLs** (e.g., 30–120s) for dynamic behavior to update promptly.
- **Don’t mix address families:** `A` answers must be IPv4; `AAAA` answers must be IPv6.
- **Serials & transfers:** Lua answers are computed at query time; changing only the Lua behavior does **not** change the zone’s SOA serial. Zone transfer and serial behavior follow normal PowerDNS rules.
- **Provider limitation:** Only the **PowerDNS** provider in DNSControl accepts `LUA()`; other providers will ignore or reject it.

## References

- PowerDNS **Lua Records** overview (syntax, examples).
- PowerDNS **Lua Reference** (functions, preset variables, objects).
- DNSControl **PowerDNS provider** page.
- DNSControl **Supported providers** table.
