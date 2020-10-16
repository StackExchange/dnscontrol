---
name: Hetzner DNS Console
title: Hetzner DNS Console
layout: default
jsId: HETZNER
---

# Hetzner DNS Console Provider

## Configuration

In your credentials file, you must provide a
[Hetzner API Key](https://dns.hetzner.com/settings/api-token).

{% highlight json %}
{
  "hetzner": {
    "api_key": "your-api-key"
  }
}
{% endhighlight %}

## Metadata

This provider does not recognize any special metadata fields unique to Hetzner
 DNS Console.

## Usage

Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none', 'NONE');
var HETZNER = NewDnsProvider("hetzner", "HETZNER");

D("example.tld", REG_NONE, DnsProvider(HETZNER),
    A("test","1.2.3.4")
);
{%endhighlight%}

## Activation

Create a new API Key in the
[Hetzner DNS Console](https://dns.hetzner.com/settings/api-token).

## Caveats

Hetzner DNS Console does not allow changing the SOA record via their API.
There is an alternative method using an import of a full BIND file, but this
 approach does not play nice with incremental changes or ignored records.
At this time you cannot update SOA records via DNSControl.
