---
name: Netcup
title: Netcup Provider
layout: default
jsId: NETCUP
---
# Netcup Provider

## Configuration
In your credentials file, you must provide your [api key, password and your customer number](https://www.netcup-wiki.de/wiki/CCP_API#Authentifizierung).

{% highlight json %}
{
  "netcup": {
    "api-key": "abc12345",
    "api-password": "abc12345",
    "customer-number": "123456"
  }
}
{% endhighlight %}

## Usage
Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none', 'NONE')
var NETCUP = NewDnsProvider('netcup' 'NETCUP');

D('example.tld', REG_NONE, DnsProvider(NETCUP),
    A('test','1.2.3.4')
);
{%endhighlight%}


## Caveats
Netcup does not allow any TTLs to be set for individual records. Thus in
the diff/preview it will always show a TTL of 0. `NS` records are also
not currently supported.
