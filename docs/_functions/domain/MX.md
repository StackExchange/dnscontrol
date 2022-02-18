---
name: MX
parameters:
  - name
  - priority
  - target
  - modifiers...
---

MX adds an MX record to the domain.

Priority should be a number.

Target should be a string representing the MX target. If it is a single label we will assume it is a relative name on the current domain. If it contains *any* dots, it should be a fully qualified domain name, ending with a `.`.

{% include startExample.html %}

```js
D("example.com", REGISTRAR, DnsProvider(R53),
  MX("@", 5, "mail"), // mx example.com -> mail.example.com
  MX("sub", 10, "mail.foo.com.")
);
```

{% include endExample.html %}
