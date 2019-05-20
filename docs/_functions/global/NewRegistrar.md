---
name: NewRegistrar
parameters:
  - name
  - type
  - meta
return: string
---

NewRegistrar registers a registrar provider. The name can be any string value you would like to use.
The type must match a valid registrar provider type identifier (see [provider page.]({{site.github.url}}/provider-list))

Metadata is an optional object, that will only be used by certain providers. See [individual provider docs]({{site.github.url}}/provider-list) for specific details.

This function will return the name as a string so that you may assign it to a variable to use inside [D](#D) directives.

{% include startExample.html %}
{% highlight js %}
var REGISTRAR = NewRegistrar("name.com", "NAMEDOTCOM");
var r53 = NewDnsProvider("R53","ROUTE53");

D("example.com", REGISTRAR, DnsProvider(r53), A("@","1.2.3.4"));
{%endhighlight%}
{% include endExample.html %}
