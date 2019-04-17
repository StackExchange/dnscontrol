---
name: DNSimple
title: DNSimple Provider
layout: default
jsId: DNSIMPLE
---
# DNSimple Provider
## Configuration
In your providers credentials file you must provide a DNSimple account access token:

{% highlight json %}
{
  "dnsimple": {
    "token": "your-dnsimple-account-access-token"
  }
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to DNSimple.

## Usage
Example Javascript:

{% highlight js %}
var REG_DNSIMPLE = NewRegistrar("dnsimple", "DNSIMPLE");
var DNSIMPLE = NewDnsProvider("dnsimple", "DNSIMPLE");

D("example.tld", REG_DNSIMPLE, DnsProvider(DNSIMPLE),
    A("test","1.2.3.4")
);
{% endhighlight %}

## Activation
DNSControl depends on a DNSimple account access token.