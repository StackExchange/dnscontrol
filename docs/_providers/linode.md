---
name: Linode
title: Linode Provider
layout: default
jsId: LINODE
---
# Linode Provider

## Configuration
In your credentials file, you must provide your
[Linode Personal Access Token](https://cloud.linode.com/profile/tokens)

{% highlight json %}
{
  "linode": {
    "token": "your-linode-personal-access-token"
  }
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to Linode.

## Usage
Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none', 'NONE')
var LINODE = NewDnsProvider("linode", "LINODE");

D("example.tld", REG_NONE, DnsProvider(LINODE),
    A("test","1.2.3.4")
);
{%endhighlight%}

## Activation
[Create Personal Access Token](https://cloud.linode.com/profile/tokens)

## Caveats
Linode does not allow all TTLs, but only a specific subset of TTLs. The following TTLs are supported
([source](https://github.com/linode/manager/blob/master/src/domains/components/SelectDNSSeconds.js)):

- 300
- 3600
- 7200
- 14400
- 28800
- 57600
- 86400
- 172800
- 345600
- 604800
- 1209600
- 2419200

The provider will automatically round up your TTL to one of these values. For example, 600 seconds would become 3600
seconds, but 300 seconds would stay 300 seconds. 
