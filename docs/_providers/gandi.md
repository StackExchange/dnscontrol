---
name: Gandi
layout: default
jsId: GANDI
---
# Gandi Provider

<<<<<<< HEAD
Gandi provides both a registrar and a dns provider implementation.
=======
Gandi provides a DnsProvider but not a Registrar.
>>>>>>> master

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

This provider does not recognize any special metadata fields unique to Gandi.

## Usage

Example javascript:

{% highlight js %}
var GANDI = NewDnsProvider("gandi", "GANDI");

D("example.tld", REG_GANDI, DnsProvider(GANDI),
    A("test","1.2.3.4")
);
{% endhighlight %}

## New domains

If a domain does not exist in your Gandi account, DNSControl
will *not* automatically add it. You'll need to do that via the
control panel manually.
