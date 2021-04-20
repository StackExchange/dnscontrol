---
name: SoftLayer DNS
title: SoftLayer DNS Provider
layout: default
jsId: SOFTLAYER
---

# SoftLayer DNS Provider

NOTE: This provider is currently has no maintainer. We are looking for
a volunteer. If this provider breaks it may be disabled or removed if
it can not be easily fixed.

## Configuration
To authenticate with SoftLayer requires at least a `username` and `api_key` for authentication. It can also optionally take a `timeout` and `endpoint_url` parameter however these are optional and will use standard defaults if not provided.

These can be supplied in the `creds.json` file:
{% highlight json %}
{
  "softlayer": {
    "username": "myusername",
    "api_key": "mysecretapikey"
  }
}
{% endhighlight %}

To maintain compatibility with existing softlayer CLI services these can also be provided by the `SL_USERNAME` and `SL_API_KEY` environment variables or specified in the `~/.softlayer`, but this is discouraged. More information about these methods can be found at [the softlayer-go library documentation](https://github.com/softlayer/softlayer-go#sessions).

## Usage
Use this provider like any other DNS Provider:

{% highlight js %}
var REG_NONE = NewRegistrar("none","NONE"); // no registrar
var SOFTLAYER = NewDnsProvider("softlayer", "SOFTLAYER");

D("example.tld", registrary, DnsProvider(SOFTLAYER),
    A("test","1.2.3.4")
);
{%endhighlight%}

## Metadata
This provider does not recognize any special metadata fields unique to SoftLayer dns.
For compatibility with the pre-generated NAMESERVER fields it's recommended to set the NS TTL to 86400 such as:

{% highlight js %}
D("example.tld", REG_NONE, DnsProvider(SOFTLAYER),
    NAMESERVER_TTL(86400),

    A("test","1.2.3.4")
);
{%endhighlight%}
