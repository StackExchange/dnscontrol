---
name: D
parameters:
  - name
  - registrar
  - modifiers...
---

`D` adds a new Domain for DNSControl to manage. The first two arguments are required: the domain name (fully qualified `example.com` without a trailing dot), and the
name of the registrar (as previously declared with [NewRegistrar](#NewRegistrar)). Any number of additional arguments may be included to add DNS Providers with [DNSProvider](#DNSProvider),
add records with [A](#A), [CNAME](#CNAME), and so forth, or add metadata.

Modifier arguments are processed according to type as follows:

- A function argument will be called with the domain object as it's only argument. Most of the [built-in modifier functions](#domain-modifiers) return such functions.
- An object argument will be merged into the domain's metadata collection.
- An array arument will have all of it's members evaluated recursively. This allows you to combine multiple common records or modifiers into a variable that can
   be used like a macro in multiple domains.

{% include startExample.html %}
{% highlight js %}
var REGISTRAR = NewRegistrar("name.com", "NAMEDOTCOM");
var r53 = NewDnsProvider("R53","ROUTE53");

// simple domain
D("example.com", REGISTRAR, DnsProvider(r53),
  A("@","1.2.3.4"),
  CNAME("test", "foo.example2.com.")
);

// "macro" for records that can be mixed into any zone
var GOOGLE_APPS_DOMAIN_MX = [
    MX('@', 1, 'aspmx.l.google.com.'),
    MX('@', 5, 'alt1.aspmx.l.google.com.'),
    MX('@', 5, 'alt2.aspmx.l.google.com.'),
    MX('@', 10, 'alt3.aspmx.l.google.com.'),
    MX('@', 10, 'alt4.aspmx.l.google.com.'),
]

D("example.com", REGISTRAR, DnsProvider(r53),
  A("@","1.2.3.4"),
  CNAME("test", "foo.example2.com."),
  GOOGLE_APPS_DOMAIN_MX
);

{%endhighlight%}
{% include endExample.html %}
