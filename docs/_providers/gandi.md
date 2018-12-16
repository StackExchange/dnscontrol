---
name: Gandi
title: Gandi Provider
layout: default
jsId: GANDI
---
# Gandi Provider

There are two providers for Gandi:

 1. `GANDI` uses the v3 API and is able to act as a registrar provider
    and a DNS provider. It is not able to handle domains that have
    migrated to the new LiveDNS API. You need to get the API key from
    the [v4 interface][].

 2. `GANDI-LIVEDNS` uses the LiveDNS API and is only able to act as a
    DNS provider. You need to get the API key from the [v5 interface][].

[v4 interface]: https://v4.gandi.net
[v5 interface]: https://v5.gandi.net

## Configuration
In your credentials file you must provide your Gandi.net API key:

{% highlight json %}
{
  "gandi": {
    "apikey": "your-gandi-key"
  }
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to Gandi.

## Usage
Example Javascript:

{% highlight js %}
var GANDI = NewDnsProvider("gandi", "GANDI");
var REG_GANDI = NewRegistrar("gandi", "GANDI");

D("example.tld", REG_GANDI, DnsProvider(GANDI),
    A("test","1.2.3.4")
);
{% endhighlight %}

## New domains
If a domain does not exist in your Gandi account, DNSControl will *not* automatically add it with the `create-domains` command. You'll need to do that via the control panel manually.


## Common errors

This is the error we see when someone uses GANDI instead of GANDI-LIVEDNS.

```
Error getting corrections: error: "Error on object : OBJECT_ZONE (CAUSE_NOTFOUND) [no such zone (id: 0)]" code: 581042
```
