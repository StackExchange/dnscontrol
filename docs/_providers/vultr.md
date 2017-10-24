---
name: Vultr
title: Vultr Provider
layout: default
jsId: VULTR
---
# Vultr Provider

## Configuration

In your providers config json file you must include a Vultr personal access token:

{% highlight json %}
{
  "vultr":{
    "token": "your-vultr-personal-access-token"
  }
}
{% endhighlight %}

## Metadata

This provider does not recognize any special metadata fields unique to Vultr.

## Usage

Example javascript:

{% highlight js %}
var VULTR = NewDnsProvider("vultr", "VULTR");

D("example.tld", REG_DNSIMPLE, DnsProvider(VULTR),
    A("test","1.2.3.4")
);
{% endhighlight %}

## Activation

Vultr depends on a Vultr personal access token.
