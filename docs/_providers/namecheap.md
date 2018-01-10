---
name: Namecheap Provider
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
  "namecheap":{
    "apikey": "yourApiKeyFromNameCheap",
    "apiuser": "yourUsername"
  }
}
{% endhighlight %}

You can optionally specify BaseURL to use a different endpoint - typically the
sandbox:

{% highlight json %}
{
  "namecheap.com":{
    "apikey": "yourApiKeyFromNameCheap",
    "apiuser": "yourUsername"
    "BaseURL": "https://api.sandbox.namecheap.com/xml.response"
  }
}
{% endhighlight %}

if BaseURL is omitted, the production namecheap url is used.


## Metadata
This provider does not recognize any special metadata fields unique to
Namecheap.

## Usage
Example Javascript:

{% highlight js %}
var REG_NAMECHEAP = NewRegistrar("namecheap","NAMECHEAP");
var R53 = NewDnsProvider("r53", "ROUTE53");

D("example.tld", REG_NAMECHEAP, DnsProvider(R53),
    A("test","1.2.3.4")
);
{%endhighlight%}

Namecheap provides custom redirect records URL, URL301, and FRAME.  These
records can be used like any other record:

{% highlight js %}
var REG_NAMECHEAP = NewRegistrar("namecheap","NAMECHEAP");
var NAMECHEAP = NewDnsProvider("namecheap","NAMECHEAP");

D("example.tld", REG_NAMECHEAP, DnsProvider(NAMECHEAP),
  URL('@', 'http://example.com/'),
  URL('www', 'http://example.com/'),
  URL301('backup', 'http://backup.example.com/')
)
{% endhighlight %}

## Activation
In order to activate API functionality on your Namecheap account, you must
enable it for your account and wait for their review process. More information
on enabling API access is [located
here](https://www.namecheap.com/support/api/intro.aspx).
