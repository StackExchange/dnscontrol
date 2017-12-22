---
name: TXT
parameters:
  - name
  - contents
  - modifiers...
---

TXT adds an TXT record To a domain. The name should be the relative
label for the record. Use `@` for the domain apex.

The contents is either a single or multiple strings.  To
specify multiple strings, include them in an array.

TXT records with multiple strings are only supported by some
providers. DNSControl will produce a validation error if the
provider does not support multiple strings.

Each string is a JavaScript string (quoted using single or double
quotes).  The (somewhat complex) quoting rules of the DNS protocol
will be done for you.

Modifers can be any number of [record modifiers](#record-modifiers) or json objects, which will be merged into the record's metadata.

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, ....,
  TXT('@', '598611146-3338560'),
  TXT('listserve', 'google-site-verification=12345'),
  TXT('multiple', ['one', 'two', 'three']),  // Multiple strings
  TXT('quoted', 'any "quotes" and escapes? ugh; no worries!'),
  TXT('_domainkey', 't=y; o=-;') // Escapes are done for you automatically.
);

{%endhighlight%}
{% include endExample.html %}
