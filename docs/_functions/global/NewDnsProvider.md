---
name: NewDnsProvider
parameters:
  - name
  - type
  - meta
return: string
---

NewDnsProvider registers a new DNS Service Provider. The name can be any string value you would like to use.
The type must match a valid dns provider type identifier (see [provider page.]({{site.github.url}}/provider-list))

Metadata is an optional object, that will only be used by certain providers. See [individual provider docs]({{site.github.url}}/provider-list) for specific details.

This function will return the name as a string so that you may assign it to a variable to use inside [D](#D) directives.

{% include startExample.html %}
{% highlight js %}
var REGISTRAR = NewRegistrar("name.com", "NAMEDOTCOM");
var R53 = NewDnsProvider("r53", "ROUTE53");

D("example.com", REGISTRAR, DnsProvider(R53), A("@","1.2.3.4"));
{%endhighlight%}
{% include endExample.html %}
