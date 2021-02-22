---
name: http.net/hosting.de
title: http.net/hosting.de Provider
layout: default
jsId: httpnet
---
# http.net/hosting.de Provider

## Configuration
In your credentials file, you must provide your [`authToken` and optionally an `ownerAccountId`](https://www.http.net/docs/api/#requests-and-authentication).

**If you want to use this provider with hosting.de or a demo system you need to provide a custom `baseURL`.**

* hosting.de: `https://secure.hosting.de`
* Demo: `https://demo.routing.net`

{% highlight json %}
{
  "http.net": {
    "authToken": "{/9Q][^~W0&MlP+T^MRF@K.=x8z<InC_X)lbvZt=Xp)&4V@W"
  },
  "hosting.de": {
    "authToken": "{/9Q][^~W0&MlP+T^MRF@K.=x8z<InC_X)lbvZt=Xp)&4V@W",
    "baseURL": "https://secure.hosting.de"
  }
}
{% endhighlight %}

## Usage
Example JavaScript:

{% highlight js %}
var REG_HTTP = NewRegistrar('http.net', 'HTTPNET')
var DNS_HTTP = NewDnsProvider('http.net' 'HTTPNET');

D('example.tld', REG_HTTP, DnsProvider(DNS_HTTP),
    A('test', '1.2.3.4')
);
{% endhighlight %}
