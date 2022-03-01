---
name: INCLUDE
parameters:
  - domain
---

Includes all records from a given domain


{% include startExample.html %}

```js
D("example.com!external", REGISTRAR, DnsProvider(R53),
  A("test", "8.8.8.8")
);

D("example.com!internal", REGISTRAR, DnsProvider(R53),
  INCLUDE("example.com!external"),
  A("home", "127.0.0.1")
);
```

{% include endExample.html %}
