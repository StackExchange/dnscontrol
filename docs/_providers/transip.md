---
name: TransIP DNS
title: TransIP DNS Provider
layout: default
jsId: TRANSIP
---

# TransIP DNS Provider

## Configuration

In your providers config json file you must include a TransIP personal access token:

{% highlight json %}
{
  "transip":{
    "AccessToken": "your-transip-personal-access-token"
  }
}
{% endhighlight %}

## Metadata

This provider does not recognize any special metadata fields unique to Vultr.

## Usage

Example javascript:

{% highlight js %}
var TRANSIP = NewDnsProvider("transip", "TRANSIP");

D("example.tld", REG_DNSIMPLE, DnsProvider(TRANSIP),
    A("test","1.2.3.4")
);
{% endhighlight %}

## Activation

TransIP depends on a TransIP personal access token.
