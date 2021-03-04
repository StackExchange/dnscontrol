---
name: hosting.de
title: hosting.de Provider
layout: default
jsId: hostingde
---
# hosting.de Provider

## Configuration
In your credentials file, you must provide your [`authToken` and optionally an `ownerAccountId`](https://www.hosting.de/api/#requests-and-authentication).

**If you want to use this provider with http.net or a demo system you need to provide a custom `baseURL`.**

* hosting.de (default): `https://secure.hosting.de`
* http.net: `https://partner.http.net`
* Demo: `https://demo.routing.net`

{% highlight json %}
{
  "hosting.de": {
    "authToken": "YOUR_API_KEY"
  },
  "http.net": {
    "authToken": "YOUR_API_KEY",
    "baseURL": "https://partner.http.net"
  }
}
{% endhighlight %}

## Usage
Example JavaScript:

{% highlight js %}
var REG_HOSTINGDE = NewRegistrar('hosting.de', 'HOSTINGDE')
var DNS_HOSTINGDE = NewDnsProvider('hosting.de' 'HOSTINGDE');

D('example.tld', REG_HOSTINGDE, DnsProvider(DNS_HOSTINGDE),
    A('test', '1.2.3.4')
);
{% endhighlight %}
