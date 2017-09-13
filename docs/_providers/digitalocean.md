---
name: Digitalocean
layout: default
jsId: DIGITALOCEAN
---
# Digitalocean Provider

## Configuration

In your providers config json file you must provide your
[Digitalocean OAuth Token](https://cloud.digitalocean.com/settings/applications)

{% highlight json %}
{
  "digitalocean":{
    "token": "your-digitalocean-ouath-token"
  }
}
{% endhighlight %}

## Metadata

This provider does not recognize any special metadata fields unique to route 53.

## Usage

Example javascript:

{% highlight js %}
var REG_NAMECOM = NewRegistrar("name.com","NAMEDOTCOM");
var DO = NewDnsProvider("do", "DIGITALOCEAN");

D("example.tld", REG_NAMECOM, DnsProvider(DO),
    A("test","1.2.3.4")
);
{%endhighlight%}

## Activation

[Create OAuth Token](https://cloud.digitalocean.com/settings/applications)
