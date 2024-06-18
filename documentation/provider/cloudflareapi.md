This is the provider for [Cloudflare](https://www.cloudflare.com/).

## Important notes

* SPF records are silently converted to RecordType `TXT` as Cloudflare API fails otherwise. See [StackExchange/dnscontrol#446](https://github.com/StackExchange/dnscontrol/issues/446).
* This provider currently fails if there are more than 1000 corrections on one domain. This only affects "push". This usually when moving a domain with many records to Cloudflare.  Try commenting out most records, then uncomment groups of 999. Typical updates are less than 1000 corrections and will not trigger this bug. See [StackExchange/dnscontrol#1440](https://github.com/StackExchange/dnscontrol/issues/1440).

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `CLOUDFLAREAPI`.

Optional fields include:

* `accountid` and `apitoken`: Authentication information
* `apikey` and `apiuser`: Old-style authentication

Example:

{% code title="creds.json" %}
```json
{
  "cloudflare": {
    "TYPE": "CLOUDFLAREAPI",
    "accountid": "your-cloudflare-account-id",
    "apitoken": "your-cloudflare-api-token"
  }
}
```
{% endcode %}

# Authentication

The Cloudflare API supports two different authentication methods.

NOTE: You can not mix the two authentication methods.  If you try, DNSControl will report an error.

## API Tokens (recommended)

The recommended (newer) method is to
provide a [Cloudflare API token](https://dash.cloudflare.com/profile/api-tokens).

This method is enabled by setting the `apitoken` value in `creds.json`:

{% code title="creds.json" %}
```json
{
  "cloudflare": {
    "TYPE": "CLOUDFLAREAPI",
    "accountid": "your-cloudflare-account-id",
    "apitoken": "your-cloudflare-api-token"
  }
}
```
{% endcode %}

* `accountid` is found in the Cloudflare portal ("Account ID") on any "Website" page.  Click on any site and you'll see the "Account ID" on the lower right side of the page.
* `apitoken` is something you must create. See [Cloudflare's documentation](https://support.cloudflare.com/hc/en-us/articles/200167836-Managing-API-Tokens-and-Keys) for instructions on how to generate and configure permissions on API tokens.  (Spoiler alert: [link](https://dash.cloudflare.com/profile/api-tokens). The token must be granted rights (authorization to do certain tasks) at a very granular level.

DNSControl requires the token to have the following permissions:

* Add: Read zones (`Zone → Zone → Read`)
* Add: Edit DNS records (`Zone → DNS → Edit`)
* Add: Enable SSL controls (`Zone → SSL and Certificates → Edit`)
* Editing Page Rules?
  * Add: Edit Page Rules (`Zone → Page Rules → Edit`)
* Creating Redirects?
  * Add: Edit Dynamic Redirect (`Zone → Dynamic Redirect → Edit`)
* Managing Cloudflare Workers? (if `manage_workers`: set to `true` or `CF_WORKER_ROUTE()` is in use.)
  * Add: Edit Worker Scripts (`Account → Workers Scripts → Edit`)
  * Add: Edit Worker Scripts (`Zone → Workers Routes → Edit`)

![Example permissions configuration](../assets/providers/cloudflareapi/example-permissions-configuration.png)

## Username+Key (not recommended)

The other (older, not recommended) method is to
provide your Cloudflare API username and access key.

This method is not recommended because these credentials give DNSControl access to everything (think of it as "super user" for your account).

This method is enabled by setting the `apikey` and `apiuser` values in `creds.json`:

{% code title="creds.json" %}
```json
{
  "cloudflare": {
    "TYPE": "CLOUDFLAREAPI",
    "accountid": "your-cloudflare-account-id",
    "apikey": "your-cloudflare-api-key",
    "apiuser": "your-cloudflare-email-address"
  }
}
```
{% endcode %}

* `accountid` (see above)
* `apiuser` is the email address associated with the account.
* `apikey` is found on [My Profile / API Tokens](https://dash.cloudflare.com/profile/api-tokens).

## Meta configuration

This provider accepts some optional metadata:

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

You can also set the default proxy mode using `DEFAULTS()` function. For example:

{% code title="dnsconfig.js" %}
```javascript
DEFAULTS(
  CF_PROXY_DEFAULT_OFF // turn proxy off when not specified otherwise
);
```
{% endcode %}

**Aliases:**

To make configuration files more readable and less prone to errors,
the following aliases are *pre-defined*:

{% code title="dnsconfig.js" %}
```javascript
// Meta settings for individual records.
var CF_PROXY_OFF = {"cloudflare_proxy": "off"};     // Proxy disabled.
var CF_PROXY_ON = {"cloudflare_proxy": "on"};       // Proxy enabled.
var CF_PROXY_FULL = {"cloudflare_proxy": "full"};   // Proxy+Railgun enabled.
// Per-domain meta settings:
// Proxy default off for entire domain (the default):
var CF_PROXY_DEFAULT_OFF = {"cloudflare_proxy_default": "off"};
// Proxy default on for entire domain:
var CF_PROXY_DEFAULT_ON = {"cloudflare_proxy_default": "on"};
// UniversalSSL off for entire domain:
var CF_UNIVERSALSSL_OFF = { cloudflare_universalssl: "off" };
// UniversalSSL on for entire domain:
var CF_UNIVERSALSSL_ON = { cloudflare_universalssl: "on" };
```
{% endcode %}

The following example shows how to set meta variables with and without aliases:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_CLOUDFLARE = NewDnsProvider("cloudflare");

D("example.com", REG_NONE, DnsProvider(DSP_CLOUDFLARE),
    A("www1","1.2.3.11", CF_PROXY_ON),        // turn proxy ON.
    A("www2","1.2.3.12", CF_PROXY_OFF),       // default is OFF, this is a no-op.
    A("www3","1.2.3.13", {"cloudflare_proxy": "on"}), // Old format.
END);
```
{% endcode %}

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_CLOUDFLARE = NewDnsProvider("cloudflare");

// Example domain where the CF proxy abides by the default (off).
D("example.com", REG_NONE, DnsProvider(DSP_CLOUDFLARE),
    A("proxied", "1.2.3.4", CF_PROXY_ON),
    A("notproxied", "1.2.3.5"),
    A("another", "1.2.3.6", CF_PROXY_ON),
    ALIAS("@", "www.example.com.", CF_PROXY_ON),
    CNAME("myalias", "www.example.com.", CF_PROXY_ON),
END);

// Example domain where the CF proxy default is set to "on":
D("example2.tld", REG_NONE, DnsProvider(DSP_CLOUDFLARE),
    CF_PROXY_DEFAULT_ON, // Enable CF proxy for all items unless otherwise noted.
    A("proxied", "1.2.3.4"),
    A("notproxied", "1.2.3.5", CF_PROXY_OFF),
    A("another", "1.2.3.6"),
    ALIAS("@", "www.example2.tld."),
    CNAME("myalias", "www.example2.tld."),
END);
```
{% endcode %}

## New domains
If a domain does not exist in your Cloudflare account, DNSControl
will automatically add it when `dnscontrol push` is executed.


## Old-style vs new-style redirects

Old-style redirects uses the [Page Rules][https://developers.cloudflare.com/rules/page-rules/] product feature, which is [going away](https://developers.cloudflare.com/rules/reference/page-rules-migration/).  In this mode,
`CF_REDIRECT` and `CF_TEMP_REDIRECT` functions generate Page Rules.

Enable it using:

```javascript
var DSP_CLOUDFLARE = NewDnsProvider("cloudflare", {
    "manage_redirects": true
});
```

New redirects uses the [Single Redirects][https://developers.cloudflare.com/rules/url-forwarding/] product feature.  In this mode, 
`CF_REDIRECT` and `CF_TEMP_REDIRECT` functions generates Single Redirects.

Enable it using:

```javascript
var DSP_CLOUDFLARE = NewDnsProvider("cloudflare", {
    "manage_single_redirects": true
});
```

{% hint style="warning" %}
New-style redirects ("Single Redirect Rules") are a new feature of DNSControl
as of v4.12.0 and may have bugs.  Please test carefully.
{% endhint %}


Conversion mode:

DNSControl can convert from old-style redirects (Page Rules) to new-style
redirect (Single Redirects). To enable this mode, set both `manage_redirects`
and `manage_single_redirects` to true.

{% hint style="warning" %}
The conversion process only handles a few, very simple, patterns.
See `providers/cloudflare/singleredirect_test.go` for a list of patterns
supported.  Please file bugs if you find problems. PRs welcome!
{% endhint %}

In conversion mode, DNSControl takes `CF_REDIRECT`/`CF_TEMP_REDIRECT`
statements and turns each of them into two records: a Page Rules and an
equivalent Single Redirects rule.

Cloudflare processes Single Redirects before Page Rules, thus it is safe to
have both at the same time, and provides an easy way to test the new-style
rules.  If they do not work properly, use the Cloudflare web-based control
panel to manually delete the new-style rule to expose the old-style rule. (and
report the bug to DNSControl!)

You'll find the new-style rule in the Cloudflare control panel.  It will have
a very long name that includes the `CF_REDIRECT`/`CF_TEMP_REDIRECT` operands
plus matcher and replacement expressions.

There is no mechanism to easily delete the old-style rules.  Either delete them
manually using the Cloudflare control panel or wait for Cloudflare to remove
the old-style Page Rule feature.

Once the conversion is complete, change
`manage_redirects` to `false` then either delete the old redirects
via the CloudFlare control panel or wait for Cloudflare to remove support for the old-style feature.

{% hint style="warning" %}
Cloudflare's announcement says that they will convert old-style redirects (Page Rules) to new-style
redirect (Single Redirects) but they do not give a date for when this will happen.  DNSControl
will probably see these new redirects as foreign and delete them.

Therefore it is probably safer to do the conversion ahead of them.

On the other hand, if you let them do the conversion, their conversion may be more correct
than DNSControl's.  However there's no way for DNSControl to manage them since the automatically-generated name will be different.

If you have suggestions on how to handle this better please file a bug.
{% endhint %}


## Redirects
The Cloudflare provider can manage "Forwarding URL" Page Rules (redirects) for your domains. Simply use the `CF_REDIRECT` and `CF_TEMP_REDIRECT` functions to make redirects:

{% code title="dnsconfig.js" %}
```javascript
// chiphacker.com should redirect to electronics.stackexchange.com

var REG_NONE = NewRegistrar("none");
var DSP_CLOUDFLARE = NewDnsProvider("cloudflare", {"manage_redirects": true}); // enable manage_redirects

D("chiphacker.com", REG_NONE, DnsProvider(DSP_CLOUDFLARE),
    // ...

    // 302 for meta subdomain
    CF_TEMP_REDIRECT("meta.chiphacker.com/*", "https://electronics.meta.stackexchange.com/$1"),

    // 301 all subdomains and preserve path
    CF_REDIRECT("*chiphacker.com/*", "https://electronics.stackexchange.com/$2"),

    // A redirect must have A records with orange cloud on. Otherwise the HTTP/HTTPS request will never arrive at Cloudflare.
    A("meta", "1.2.3.4", CF_PROXY_ON),

    // ...
END);
```
{% endcode %}

Notice a few details:

1. We need an A record with cloudflare proxy on, or the page rule will never run.
2. The IP address in those A records may be mostly irrelevant, as cloudflare should handle all requests (assuming some page rule matches).
3. Ordering matters for priority. CF_REDIRECT records will be added in the order they appear in your js. So put catch-alls at the bottom.
4. if _any_ `CF_REDIRECT` or `CF_TEMP_REDIRECT` functions are used then `dnscontrol` will manage _all_ "Forwarding URL" type Page Rules for the domain. Page Rule types other than "Forwarding URL" will be left alone. In other words, `dnscontrol` will delete any Forwarding URL it doesn't recognize. Be careful!

## Worker routes
The Cloudflare provider can manage Worker Routes for your domains. Simply use the `CF_WORKER_ROUTE` function passing the route pattern and the worker name:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_CLOUDFLARE = NewDnsProvider("cloudflare", {"manage_workers": true}); // enable managing worker routes

D("foo.com", REG_NONE, DnsProvider(DSP_CLOUDFLARE),
    // Assign the patterns `api.foo.com/*` and `foo.com/api/*` to `my-worker` script.
    CF_WORKER_ROUTE("api.foo.com/*", "my-worker"),
    CF_WORKER_ROUTE("foo.com/api/*", "my-worker"),
END);
```
{% endcode %}

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

```shell
cd integrationTest              # NOTE: Not needed if already in that subdirectory
go test -v -verbose -provider CLOUDFLAREAPI -cfworkers=false
```

When `-cfworkers=false` is set, tests related to Workers are skipped.  The Account ID is not required.


## Cloudflare special TTLs

Cloudflare plays tricks with TTLs.  Cloudflare uses "1" to mean "auto-ttl";
which as far as we can tell means 300 seconds (5 minutes) with the option that
CloudFlare may dynamically adjust the actual TTL. In the Cloudflare API,
setting the TTL to 300 results in the TTL being set to 1.

If the TTL isn't set to 1, Cloudflare has a minimum of 1 minutes.

A TTL of 0 tells DNSControl to use the default TTL for that provider, which is 1.

In summary:
* TTL of 0, 1 and 300 are all the same ("auto TTL").
* TTL of 2-60 are all the same as 60.
* TTL of 61-299, and 301 to infinity are not magic.

Some of this is documented on the Cloudflare website's [Time to Live (TTL)](https://developers.cloudflare.com/dns/manage-dns-records/reference/ttl/) page.
