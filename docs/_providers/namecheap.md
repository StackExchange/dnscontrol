---
name: "Namecheap"
layout: default
jsId: NAMECHEAP
---

# Namecheap Provider

Namecheap only provides a registrar provider implementation.

## Configuration

In your providers config json file you must provide your Namecheap api
username and key:

{% highlight json %}
{
  "namecheap.com":{
    "apikey": "yourApiKeyFromNameCheap",
    "apiuser": "yourUsername"
  }
}
{% endhighlight %}

## Metadata

This provider does not recognize any special metadata fields unique to
Namecheap.

## Usage

Example javascript:

{% highlight js %}
var namecheap = NewRegistrar("namecheap.com","NAMECHEAP");
var R53 = NewDnsProvider("r53", ROUTE53);

D("example.tld", namecheap, DnsProvider(R53),
    A("test","1.2.3.4")
);
{%endhighlight%}

## Activation

In order to activate api functionality on your Namecheap account, you must
enable it for your account and wait for their review process. More information
on enabling API access is [located
here](https://www.namecheap.com/support/api/intro.aspx).
