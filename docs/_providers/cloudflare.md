---
name: Cloudflare
title: Cloudflare Provider
layout: default
jsId: CLOUDFLAREAPI
---
# Cloudflare Provider

This is the provider for [Cloudflare](https://www.cloudflare.com/).

## Important notes

* When using `SPF()` or the `SPF_BUILDER()` the records are converted to RecordType `TXT` as Cloudflare API fails otherwise. See more [here](https://github.com/StackExchange/dnscontrol/issues/446).

## Configuration

The Cloudflare API supports two different authentication methods.

The recommended (newer) method is to
provide a [Cloudflare API token](https://dash.cloudflare.com/profile/api-tokens).

This method is enabled by setting the "apitoken" value in `creds.json`:

{% highlight json %}
{
  "cloudflare": {
    "apitoken": "your-cloudflare-api-token",
    "accountid": "your-cloudflare-account-id"
  }
}
{% endhighlight %}

See [Cloudflare's documentation](https://support.cloudflare.com/hc/en-us/articles/200167836-Managing-API-Tokens-and-Keys) for instructions on how to generate and configure permissions on API tokens.

A token can be granted rights (authorization to do certain tasks) at a very granular level.  DNSControl requires the token to have the following rights:

* Read zones (`Zone → Zone → Read`)
* Edit DNS records (`Zone → DNS → Edit`)
* Edit Page Rules (`Zone → Page Rules → Edit`) (Only required if `manage_redirects` is true for any dommain.)
* Enable SSL controls (`Zone → SSL and Certificates → Edit`)
* If Cloudflare Workers are being managed: (if `manage_workers`: set to `true` or `CF_WORKER_ROUTE()` is in use.)
  * Edit Worker Scripts (`Account → Workers Scripts → Edit`)
  * Edit Worker Scripts (`Zone → Workers Routes → Edit`)
* FYI: [An example permissions configuration](https://user-images.githubusercontent.com/210250/136301050-1fd430bf-21b6-428b-aa54-f6009964031d.png)

The other (older, not recommended) method is to
provide your Cloudflare API username and access key.
This key is available under "My Settings".

This method is not recommended because these credentials give DNSControl access to the entire Cloudflare API.

This method is enabled by setting the "apikey" and "apiuser" values in `creds.json`:

{% highlight json %}
{
  "cloudflare": {
    "apikey": "your-cloudflare-api-key",
    "apiuser": "your-cloudflare-email-address",
    "accountid": "your-cloudflare-account-id"
  }
}
{% endhighlight %}

You can not mix `apikey/apiuser` and `apitoken`.  If all three values are set, you will receive an error.

You should also set the "accountid" value.  This is optional but may become required some day therefore we recommend setting it.
The Account ID is used to disambiguate when API key has access to multiple Cloudflare accounts. For example, when creating domains this key is used to determine which account to place the new domain.  It is also required when using Workers.

The "accountid" is found in the Cloudflare portal ("Account ID") on the DNS page. Set it in `creds.json`:

{% highlight json %}
{
  "cloudflare": {
    "apitoken": "...",
    "accountid": "your-cloudflare-account-id",
  }
}
{% endhighlight %}

Older `creds.json` files that do not have accountid set may work for now, but not in the future.

## Metadata
Record level metadata available:
   * `cloudflare_proxy` ("on", "off", or "full")

Domain level metadata available:
   * `cloudflare_proxy_default` ("on", "off", or "full")
   * `cloudflare_universalssl` (unset to leave this setting unmanaged; otherwise use "on" or "off")
     * NOTE: If "universal SSL" isn't working, verify the API key has `Zone → SSL and Certificates → Edit` permissions. See above.

Provider level metadata available:
   * `ip_conversions`
   * `manage_redirects`: set to `true` to manage page-rule based redirects
   * `manage_workers`: set to `true` to manage cloud workers (`CF_WORKER_ROUTE`)

What does on/off/full mean?

   * "off" disables the Cloudflare proxy
   * "on" enables the Cloudflare proxy (turns on the "orange cloud")
   * "full" is the same as "on" but also enables Railgun.  DNSControl will prevent you from accidentally enabling "full" on a CNAME that points to an A record that is set to "off", as this is generally not desired.

You can also set the default proxy mode using `DEFAULTS()` function, see:
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
    A('www3','1.2.3.13', {'cloudflare_proxy': 'on'}) // Old format.
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


## New domains
If a domain does not exist in your Cloudflare account, DNSControl
will *not* automatically add it. You'll need to do that via the
control panel manually or via the `dnscontrol create-domains` command.


## Redirects
The Cloudflare provider can manage "Forwarding URL" Page Rules (redirects) for your domains. Simply use the `CF_REDIRECT` and `CF_TEMP_REDIRECT` functions to make redirects:

{% highlight js %}

// chiphacker.com should redirect to electronics.stackexchange.com

var CLOUDFLARE = NewDnsProvider('cloudflare','CLOUDFLAREAPI', {"manage_redirects": true}); // enable manage_redirects

D("chiphacker.com", REG_NONE, DnsProvider(CLOUDFLARE),
    // ...

    // 302 for meta subdomain
    CF_TEMP_REDIRECT("meta.chiphacker.com/*", "https://electronics.meta.stackexchange.com/$1"),

    // 301 all subdomains and preserve path
    CF_REDIRECT("*chiphacker.com/*", "https://electronics.stackexchange.com/$2"),

    // A redirect must have A records with orange cloud on. Otherwise the HTTP/HTTPS request will never arrive at Cloudflare.
    A("meta", "1.2.3.4", CF_PROXY_ON),

    // ...
);
{%endhighlight%}

Notice a few details:

1. We need an A record with cloudflare proxy on, or the page rule will never run.
2. The IP address in those A records may be mostly irrelevant, as cloudflare should handle all requests (assuming some page rule matches).
3. Ordering matters for priority. CF_REDIRECT records will be added in the order they appear in your js. So put catch-alls at the bottom.
4. if _any_ `CF_REDIRECT` or `CF_TEMP_REDIRECT` functions are used then `dnscontrol` will manage _all_ "Forwarding URL" type Page Rules for the domain. Page Rule types other than "Forwarding URL” will be left alone. In other words, `dnscontrol` will delete any Forwarding URL it doesn't recognize. Be careful!

## Worker routes
The Cloudflare provider can manage Worker Routes for your domains. Simply use the `CF_WORKER_ROUTE` function passing the route pattern and the worker name:

{% highlight js %}

var CLOUDFLARE = NewDnsProvider('cloudflare','CLOUDFLAREAPI', {"manage_workers": true}); // enable managing worker routes

D("foo.com", REG_NONE, DnsProvider(CLOUDFLARE),
    // Assign the patterns `api.foo.com/*` and `foo.com/api/*` to `my-worker` script.
    CF_WORKER_ROUTE("api.foo.com/*", "my-worker"),
    CF_WORKER_ROUTE("foo.com/api/*", "my-worker"),
);

{%endhighlight%}

The API key you use must be enabled to edit workers.  In the portal, edit the API key,
under "Permissions" add "Account", "Workers Scripts", "Edit". Without this permission you may see errors that mention "failed fetching worker route list from cloudflare: bad status code from cloudflare: 403 not 200"

Please notice that if _any_ `CF_WORKER_ROUTE` function is used then `dnscontrol` will manage _all_
Worker Routes for the domain. To be clear: this means it will delete existing routes that
were created outside of DNSControl.

## Integration testing

The integration tests assume that Cloudflare Workers are enabled and the credentials used
have the required permissions listed above.  The flag `-cfworkers=false` will disable tests related to Workers.
This flag is intended for use with legacy domains where the integration test credentials do not
have access to read/edit Workers. This flag will eventually go away.

{% highlight bash %}

go test -v -verbose -provider CLOUDFLAREAPI -cfworkers=false

{%endhighlight%}

When `-cfworkers=false` is set, tests related to Workers are skipped.  The Account ID is not required.
