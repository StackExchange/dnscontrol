---
name: CNAME
parameters:
  - name
  - target
  - modifiers...
---

CNAME adds a CNAME record to the domain. The name should be the relative label for the domain.
Using `@` or `*` for CNAME records is not recommended, as different providers support them differently.

Target should be a string representing the CNAME target. If it is a single label we will assume it is a relative name on the current domain. If it contains *any* dots, it should be a fully qualified domain name, ending with a `.`.

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, DnsProvider("R53"),
  CNAME("foo", "google.com."), // foo.example.com -> google.com
  CNAME("abc", "@"), // abc.example.com -> example.com
  CNAME("def", "test"), // def.example.com -> test.example.com
);

{%endhighlight%}
{% include endExample.html %}
