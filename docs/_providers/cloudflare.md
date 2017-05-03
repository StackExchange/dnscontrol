---
name: Cloudflare
layout: default
jsId: CLOUDFLAREAPI
---
# Cloudflare Provider

## Configuration

In your providers config json file you must provide your cloudflare.com api
username and access token:

{% highlight json %}
{
  "cloudflare.com":{
    "apikey": "your-cloudflare-api-key",
    "apiuser": "your-cloudflare-email-address"
  }
}
{% endhighlight %}

## Metadata

Record level metadata availible:
   * cloudflare_proxy ("on", "off", or "full")

Domain level metadata availible:
   * cloudflare_proxy_default ("on", "off", or "full")

Provider level metadata availible:
   * ip_conversions

## Usage

Example javascript:

{% highlight js %}
var REG_NAMECOM = NewRegistrar("name.com","NAMEDOTCOM");
var CFLARE = NewDnsProvider("cloudflare.com","CLOUDFLAREAPI");

D("example.tld", REG_NAMECOM, DnsProvider(CFLARE),
    A("test","1.2.3.4"),
    A("www","1.2.3.4", {cloudlfare_proxy:"on"}),
    ALIAS("@","test.example.tld",{cloudflare_proxy:"on"})
);
{%endhighlight%}

## Activation

DNSControl depends on a Cloudflare Global API Key that's available under "My Settings".

**Note** the domain must be in Cloudflare already before it can be worked on. DNSControl will not create new domains (currently) before working on them. 
