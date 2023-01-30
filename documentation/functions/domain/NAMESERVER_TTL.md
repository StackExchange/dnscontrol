---
name: NAMESERVER_TTL
parameters:
  - ttl
parameter_types:
  ttl: Duration
  target: string
  modifiers...: RecordModifier[]
---

NAMESERVER_TTL sets the TTL on the domain apex NS RRs defined by [NAMESERVER](NAMESERVER.md).

The value can be an integer or a string. See [TTL](../record/TTL.md) for examples.

```javascript
D('example.com', REGISTRAR, DnsProvider('R53'),
  NAMESERVER_TTL('2d'),
  NAMESERVER('ns')
);
```
