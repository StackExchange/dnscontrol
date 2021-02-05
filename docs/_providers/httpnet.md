---
name: http.net
title: http.net Provider
layout: default
jsId: httpnet
---
# http.net Provider

## Configuration
In your credentials file, you must provide your [authToken and optionally an ownerAccountId](https://www.http.net/docs/api/#requests-and-authentication).

{% highlight json %}
{
  "http.net": {
    "authToken": "{/9Q][^~W0&MlP+T^MRF@K.=x8z<InC_X)lbvZt=Xp)&4V@W"
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
