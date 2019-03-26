---
name: OpenSRS
layout: default
jsId: OPENSRS
---
# OpenSRS Provider

## Configuration

In your providers config json file you must provide an OpenSRS api key, username, and (optionally?) baseurl:

{% highlight json %}
{
  "opensrs":{
    "apikey": "your api key",
    "username": "your username",
    "baseurl": "the base url for API calls"
  }
}
{% endhighlight %}

FILL IN how to get an API key and username.

## Metadata

This provider does not recognize any special metadata fields unique to OpenSRS.

## Usage

Example javascript:

Example javascript (DNS hosted with OpenSRS):
{% highlight js %}
var REG_OPENSRS = NewRegistrar("opensrs", "OPENSRS");
var OPENSRS = NewDnsProvider("opensrs", "OPENSRS");

D("example.tld", REG_OPENSRS, DnsProvider(OPENSRS),
    A("test","1.2.3.4")
);
{% endhighlight %}

Example javascript (Registrar only. DNS hosted elsewhere):

{% highlight js %}
var REG_OPENSRS = NewRegistrar("ovh", "OPENSRS");
var R53 = NewDnsProvider("r53", "ROUTE53");

D("example.tld", REG_OPENSRS, DnsProvider(R53),
    A("test","1.2.3.4")
);
{%endhighlight%}


## Activation

To obtain the OpenSRS keys, ... (PLEASE FILL IN)
