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
   * `cloudflare_proxy` ("on", "off", or "full")

Domain level metadata availible:
   * `cloudflare_proxy_default` ("on", "off", or "full")

Provider level metadata availible:
   * `ip_conversions`
   * `manage_redirects`: set to `true` to manage page-rule based redirects

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

## Redirects

The cloudflare provider can manage Page-Rule based redirects for your domains. Simply use the `CF_REDIRECT` and `CF_TEMP_REDIRECT` functions to make redirects:

{% highlight js %}

// chiphacker.com is an alias for electronics.stackexchange.com

D("chiphacker.com", REG_NAMECOM, DnsProvider(CFLARE),
    // must have A records with orange cloud on. Otherwise page rule will never run.
    A("@","1.2.3.4", CF_PROXY_ON),
    A("www", "1.2.3.4", CF_PROXY_ON)
    A("meta", "1.2.3.4", CF_PROXY_ON),

    // 302 for meta subdomain
    CF_TEMP_REDIRECT("meta.chiphacker.com/*", "https://electronics.meta.stackexchange.com/$1),

    // 301 all subdomains and preserve path
    CF_REDIRECT("*chiphacker.com/*", "https://electronics.stackexchange.com/$2),
);
{%endhighlight%}

Notice a few details:

1. We need an A record with cloudflare proxy on, or the page rule will never run. 
2. The IP address in those A records may be mostly irrelevant, as cloudflare should handle all requests (assuming some page rule matches).
3. Ordering matters for priority. CF_REDIRECT records will be added in the order they appear in your js. So put catch-alls at the bottom.