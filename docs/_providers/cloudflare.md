---
name: Cloudflare
title: Cloudflare Provider
layout: default
jsId: CLOUDFLAREAPI
---
# Cloudflare Provider

## Important notes

* When using `SPF()` or the `SPF_BUILDER()` the records are converted to RecordType `TXT` as Cloudflare API fails otherwise. See more [here](https://github.com/StackExchange/dnscontrol/issues/446).

## Configuration
In the credentials file you must provide a [Cloudflare API token](https://dash.cloudflare.com/profile/api-tokens):

{% highlight json %}
{
  "cloudflare": {
    "apitoken": "your-cloudflare-api-token"
  }
}
{% endhighlight %}

Make sure the token has at least the right read zones and edit DNS records (i.e. `Zone → Zone → Read` and `Zone → DNS → Edit`; to modify Page Rules additionally requires `Zone → Page Rules → Edit`);
checkout [Cloudflare's documentation](https://support.cloudflare.com/hc/en-us/articles/200167836-Managing-API-Tokens-and-Keys) for instructions on how to generate and configure permissions on API tokens.


Or you can provide your Cloudflare API username and access key instead (but it isn't recommended because those credentials give DNSControl access to the complete Cloudflare API):

{% highlight json %}
{
  "cloudflare": {
    "apikey": "your-cloudflare-api-key",
    "apiuser": "your-cloudflare-email-address"
  }
}
{% endhighlight %}

If your Cloudflare account has access to multiple Cloudflare accounts, you can specify which Cloudflare account should be used when adding new domains:

{% highlight json %}
{
  "cloudflare": {
    "apitoken": "...",
    "accountid": "your-cloudflare-account-id",
    "accountname": "your-cloudflare-account-name"
  }
}
{% endhighlight %}

## Metadata
Record level metadata available:
   * `cloudflare_proxy` ("on", "off", or "full")

Domain level metadata available:
   * `cloudflare_proxy_default` ("on", "off", or "full")
   * `cloudflare_universalssl` (unset to keep untouched; otherwise "on, or "off")

Provider level metadata available:
   * `ip_conversions`
   * `manage_redirects`: set to `true` to manage page-rule based redirects

What does on/off/full mean?

   * "off" disables the Cloudflare proxy
   * "on" enables the Cloudflare proxy (turns on the "orange cloud")
   * "full" is the same as "on" but also enables Railgun.  DNSControl will prevent you from accidentally enabling "full" on a CNAME that points to an A record that is set to "off", as this is generally not desired.

Good to know: You can also set the default proxy mode using `DEFAULTS()` function, see:
{% highlight js %}

DEFAULTS(
	CF_PROXY_DEFAULT_OFF // turn proxy off when not specified otherwise
);

{% endhighlight %}

**Aliases:**

To make configuration files more readable and less prone to errors,
the following aliases are *pre-defined*:

{% highlight js %}
// Meta settings for individual records.
var CF_PROXY_OFF = {'cloudflare_proxy': 'off'};     // Proxy disabled.
var CF_PROXY_ON = {'cloudflare_proxy': 'on'};       // Proxy enabled.
var CF_PROXY_FULL = {'cloudflare_proxy': 'full'};   // Proxy+Railgun enabled.
// Per-domain meta settings:
// Proxy default off for entire domain (the default):
var CF_PROXY_DEFAULT_OFF = {'cloudflare_proxy_default': 'off'};
// Proxy default on for entire domain:
var CF_PROXY_DEFAULT_ON = {'cloudflare_proxy_default': 'on'};
// UniversalSSL off for entire domain:
var CF_UNIVERSALSSL_OFF = { cloudflare_universalssl: 'off' };
// UniversalSSL on for entire domain:
var CF_UNIVERSALSSL_ON = { cloudflare_universalssl: 'on' };
{% endhighlight %}

The following example shows how to set meta variables with and without aliases:

{% highlight js %}
D('example.tld', REG_NONE, DnsProvider(CLOUDFLARE),
    A('www1','1.2.3.11', CF_PROXY_ON),        // turn proxy ON.
    A('www2','1.2.3.12', CF_PROXY_OFF),       // default is OFF, this is a no-op.
    A('www3','1.2.3.13', {'cloudflare_proxy': 'on'}) // why would anyone do this?
);
{% endhighlight %}

## Usage
Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none', 'NONE')
var CLOUDFLARE = NewDnsProvider('cloudflare','CLOUDFLAREAPI');

// Example domain where the CF proxy abides by the default (off).
D('example.tld', REG_NONE, DnsProvider(CLOUDFLARE),
    A('proxied','1.2.3.4', CF_PROXY_ON),
    A('notproxied','1.2.3.5'),
    A('another','1.2.3.6', CF_PROXY_ON),
    ALIAS('@','www.example.tld.', CF_PROXY_ON),
    CNAME('myalias','www.example.tld.', CF_PROXY_ON)
);

// Example domain where the CF proxy default is set to "on":
D('example2.tld', REG_NONE, DnsProvider(CLOUDFLARE),
    CF_PROXY_DEFAULT_ON, // Enable CF proxy for all items unless otherwise noted.
    A('proxied','1.2.3.4'),
    A('notproxied','1.2.3.5', CF_PROXY_OFF),
    A('another','1.2.3.6'),
    ALIAS('@','www.example2.tld.'),
    CNAME('myalias','www.example2.tld.')
);
{%endhighlight%}

## Activation
DNSControl depends on a Cloudflare Global API Key that's available under "My Settings".

## New domains
If a domain does not exist in your Cloudflare account, DNSControl
will *not* automatically add it. You'll need to do that via the
control panel manually or via the `dnscontrol create-domains` command.

## Redirects
The Cloudflare provider can manage "Forwarding URL" Page Rules (redirects) for your domains. Simply use the `CF_REDIRECT` and `CF_TEMP_REDIRECT` functions to make redirects:

{% highlight js %}

// chiphacker.com is an alias for electronics.stackexchange.com

var CLOUDFLARE = NewDnsProvider('cloudflare','CLOUDFLAREAPI', {"manage_redirects": true}); // enable manage_redirects

D("chiphacker.com", REG_NONE, DnsProvider(CLOUDFLARE),
    // must have A records with orange cloud on. Otherwise page rule will never run.
    A("@","1.2.3.4", CF_PROXY_ON),
    A("www", "1.2.3.4", CF_PROXY_ON)
    A("meta", "1.2.3.4", CF_PROXY_ON),

    // 302 for meta subdomain
    CF_TEMP_REDIRECT("meta.chiphacker.com/*", "https://electronics.meta.stackexchange.com/$1"),

    // 301 all subdomains and preserve path
    CF_REDIRECT("*chiphacker.com/*", "https://electronics.stackexchange.com/$2"),
);
{%endhighlight%}

Notice a few details:

1. We need an A record with cloudflare proxy on, or the page rule will never run.
2. The IP address in those A records may be mostly irrelevant, as cloudflare should handle all requests (assuming some page rule matches).
3. Ordering matters for priority. CF_REDIRECT records will be added in the order they appear in your js. So put catch-alls at the bottom.
4. if _any_ `CF_REDIRECT` or `CF_TEMP_REDIRECT` functions are used then `dnscontrol` will manage _all_ "Forwarding URL" type Page Rules for the domain. Page Rule types other than "Forwarding URL” will be left alone.
