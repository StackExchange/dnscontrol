---
name: Gandi
layout: default
jsId: GANDI
---
# Gandi Provider

Gandi provides both a registrar and a dns provider implementation.

## Configuration

In your providers config json file you must provide your Gandi.net api key:

{% highlight json %}
{
  "gandi":{
    "apikey": "your-gandi-key"
  }
}
{% endhighlight %}

## Metadata

This provider does not recognize any special metadata fields unique to DNSimple.

## Usage

Example javascript:

{% highlight js %}
var REG_GANDI = NewRegistrar("gandi", "GANDI");
var GANDI = NewDnsProvider("gandi", "GANDI");

D("example.tld", REG_GANDI, DnsProvider(GANDI),
    A("test","1.2.3.4")
);
{% endhighlight %}
