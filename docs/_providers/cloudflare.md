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
    "apiuser": "your-cloudflare-username"
  }
}
{% endhighlight %}

## Metadata

Record level metadata availible:
   * cloudflare_proxy ("true" or "false")

Domain level metadata availible:
   * cloudflare_proxy_default ("true" or "false")

Provider level metadata availible:
   * ip_conversions

## Usage

Example javascript:

{% highlight js %}
var REG_NAMECOM = NewRegistrar("name.com","NAMEDOTCOM");
var CFLARE = NewDnsProvider("cloudflare.com","CLOUDFLAREAPI");

D("example.tld", REG_NAMECOM, DnsProvider(CFLARE),
    A("test","1.2.3.4")
);
{%endhighlight%}

## Activation

DNSControl depends on a Cloudflare Global API Key that's available under "My Settings".
