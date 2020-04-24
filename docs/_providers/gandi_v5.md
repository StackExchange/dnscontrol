---
name: Gandi_v5
title: Gandi_v5 Provider
layout: default
jsId: GANDI_V5
---
# Gandi_v5 Provider

`GANDI_V5` uses the v5 API and can act as a registrar provider
    or a DNS provider. It is only able to work with domains
    migrated to the new LiveDNS API, which should be all domains.
    API keys are assigned to particular users.  Go to User Settings,
    "Manage the user account and security settings", the "Security"
    tab, then regenerate the "Production API key".

* API Documentation: https://api.gandi.net/docs
* API Endpoint: https://api.gandi.net/

## Configuration
In your credentials file you must provide your Gandi.net API key.
The [sharing_id](https://api.gandi.net/docs/reference/) is optional.

The `sharing_id` selects between different organizations which your account is
a member of; to manage domains in multiple organizations, you can use multiple
`creds.json` entries.  The first parameter to `NewDnsProvider` is the key to
use in `creds.json`, and you can register multiple configured providers on the
same backend `"GANDI_V5"` provider.
(NB: in practice, this doesn't appear to be necessary and `sharing_id` is not
enforced?)

{% highlight json %}
{
  "gandi": {
    "apikey": "your-gandi-key",
    "sharing_id": "your-sharing_id"
  }
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to Gandi.

## Limitations
This provider does not support using `ALIAS` in combination with DNSSEC,
whether `AUTODNSSEC` or otherwise.

This provider only supports `ALIAS` on the `"@"` zone apex, not on any other
names.

## Usage
Example Javascript:

{% highlight js %}
var GANDI = NewDnsProvider("gandi", "GANDI_V5");
var REG_GANDI = NewRegistrar("gandi", "GANDI_V5");

D("example.tld", REG_GANDI, DnsProvider(GANDI),
    A("test","1.2.3.4")
);
{% endhighlight %}

If you are converting from the old "GANDI" provider,
simply change "GANDI" to "GANDI_V5" in `dnsconfig.js`.
Be sure to test with `dnscontrol preview` before running `dnscontrol push`.

## New domains
If a domain does not exist in your Gandi account, DNSControl will *not* automatically add it with the `create-domains` command. You'll need to do that via the web UI manually.


## Common errors

This is the error you'll see if your API key is invalid.

```
Error getting corrections: 401: The server could not verify that you authorized to access the document you requested. Either you supplied the wrong credentials (e.g., bad api key), or your access token has expired
```
