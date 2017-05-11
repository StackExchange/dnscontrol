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

Note: Aliases are pre-defined as follows:

{% highlight json %}
var CF_PROXY_OFF = {'cloudflare_proxy': 'off'};     // Default/off.
var CF_PROXY_ON = {'cloudflare_proxy': 'on'};       // Sites safe to proxy.
var CF_PROXY_FULL = {'cloudflare_proxy': 'full'};   // Sites safe to railgun.
var SET_PROXY_DEFAULT_TRUE = CF_PROXY_ON; // Turn on CF proxy for entire domain.
var SET_PROXY_DEFAULT_FALSE = CF_PROXY_OFF; // basically a no-op.
{% endhighlight %}

Thus metadata items can be used in a more readable way:

{% highlight json %}
D("example.tld", REG_NAMECOM, DnsProvider(CFLARE),
    A("www1","1.2.3.11", CF_PROXY_ON),
    A("www2","1.2.3.12", CF_PROXY_OFF), // default is OFF, this is a no-op.
);
{% endhighlight %}

or simply:

{% highlight json %}
D("example.tld", REG_NAMECOM, DnsProvider(CFLARE),
    SET_PROXY_DEFAULT_TRUE,  // Enable CF proxy for all items:
    A("www1","1.2.3.11"),
    A("www2","1.2.3.12"),
    A("www3","1.2.3.13", CF_PROXY_OFF),  // Except this one!
);
{% endhighlight %}


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

## New domains

If a domain does not exist in your CloudFlare account, DNSControl
will *not* automatically add it. You'll need to do that via the
control panel manually or via the command `dnscontrol create-domains`
-command.
