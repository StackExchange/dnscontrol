---
name: Gandi
title: Gandi Provider
layout: default
jsId: GANDI
---
# Gandi Provider

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
