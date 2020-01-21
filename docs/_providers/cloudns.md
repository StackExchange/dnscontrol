---
name: ClouDNS
title: ClouDNS Provider
layout: default
jsId: CLOUDNS
---
# ClouDNS Provider

## Configuration
In your credentials file, you must provide your [Api user ID and password](https://asia.cloudns.net/wiki/article/42/). 

Current version of provider doesn't support `sub-auth-id` or  `sub-auth-user`. 

{% highlight json %}
{
  "cloudns": {
    "auth-id": "12345",
    "auth-password": "your-password"
  }
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to ClouDNS.

## Usage
Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none', 'NONE')
var CLOUDNS = NewDnsProvider("cloudns", "CLOUDNS");

D("example.tld", REG_NONE, DnsProvider(CLOUDNS),
    A("test","1.2.3.4")
);
{%endhighlight%}

## Activation
[Create Auth ID](https://asia.cloudns.net/api-settings/).  Only paid account can use API

## Caveats
ClouDNS does not allow all TTLs, but only a specific subset of TTLs. The following [TTLs are supported](https://asia.cloudns.net/wiki/article/188/):
- 60  (1 minute)
- 300 (5 minutes)
- 900 (15 minutes)
- 1800 (30 minutes)
- 3600 (1 hour)
- 21600 (6 hours)
- 43200 (12 hours)
- 86400 (1 day)
- 172800 (2 days)
- 259200 (3 days)
- 604800 (1 week)
- 1209600 (2 weeks)
- 2419200 (4 weeks)

The provider will automatically round up your TTL to one of these values. For example, 350 seconds would become 900
seconds, but 300 seconds would stay 300 seconds. 
