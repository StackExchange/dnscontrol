---
name: A
parameters:
  - name
  - address
  - modifiers...
---

A adds an A record To a domain. The name should be the relative label for the record. Use `@` for the domain apex.

The address should be an ip address, either a string, or a numeric value obtained via [IP](#IP).

Modifiers can be any number of [record modifiers](#record-modifiers) or json objects, which will be merged into the record's metadata.

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, DnsProvider("R53"),
  A("@", "1.2.3.4"),
  A("foo", "2.3.4.5"),
  A("test.foo", IP("1.2.3.4"), TTL(5000)),
  A("*", "1.2.3.4", {foo: 42})
);

{%endhighlight%}
{% include endExample.html %}