---
name: NS
parameters:
  - name
  - target
  - modifiers...
---

NS adds a NS record to the domain. The name should be the relative label for the domain.

The name may not be `@` (the bare domain), as that is controlled via `NAMESERVER()`.
The difference between `NS()` and `NAMESERVER()` is explained in the `NAMESERVER()` description.


Target should be a string representing the NS target. If it is a single label we will assume it is a relative name on the current domain. If it contains *any* dots, it should be a fully qualified domain name, ending with a `.`.

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, DnsProvider("R53"),
  NS("foo", "ns1.example2.com."), // Delegate ".foo.example.com" zone to another server.
  NS("foo", "ns2.example2.com."), // Delegate ".foo.example.com" zone to another server.
  A("ns1.example2.com", "10.10.10.10"), // Glue records
  A("ns2.example2.com", "10.10.10.20"), // Glue records
);

{%endhighlight%}
{% include endExample.html %}
