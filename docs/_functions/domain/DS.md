---
name: DS
parameters:
  - name
  - keytag
  - algorithm
  - digesttype
  - digest
  - modifiers...
---

DS adds a DS record to the domain.

Key Tag should be a number.

Algorithm should be a number.

Digest Type must be a number.

Digest must be a string.

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, DnsProvider(R53),
  DS("example.com", 2371, 13, 2, "ABCDEF")
);

{%endhighlight%}
{% include endExample.html %}