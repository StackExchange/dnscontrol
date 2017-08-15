---
name: TXT
parameters:
  - name
  - contents
  - modifiers...
---

TXT adds an TXT record To a domain. The name should be the relative
label for the record. Use `@` for the domain apex.

The contents is a single string.  While DNS permits multiple
strings in TXT records, that is not supported at this time.

The string is a JavaScript string (quoted using single or double
quotes).  The (somewhat complex) quoting rules of the DNS protocol
will be done for you.

Modifers can be any number of [record modifiers](#record-modifiers) or json objects, which will be merged into the record's metadata.

{% include startExample.html %}
{% highlight js %}

D("example.com", REGISTRAR, ....,
  TXT('@', '598611146-3338560'),
  TXT('listserve', 'google-site-verification=12345'),
);

{%endhighlight%}
{% include endExample.html %}
