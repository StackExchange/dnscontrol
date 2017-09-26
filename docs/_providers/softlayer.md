---
name: "SoftLayer DNS"
layout: default
jsId: SOFTLAYER
---

# SoftLayer DNS Provider

## Configuration

To authenticate with softlayer requires at least a `username` and `api_key` for authentication.
It can also optionally take a `timeout` and `endpoint_url` parameter however these are optional and will use standard defaults if not provided.
These can be supplied via the standard 'creds.json' like so:
{% highlight json %}
    "softlayer": {
        "username": "myusername",
        "api_key": "mysecretapikey"
    }
{% endhighlight %}

To maintain compatibility with existing softlayer CLI services these can also be provided by the `SL_USERNAME` and `SL_API_KEY` environment variables or specified in the ~/.softlayer.
More information about these methods can be found at [the softlayer-go library documentation](https://github.com/softlayer/softlayer-go#sessions).

## Usage

Use this provider like any other DNS Provider:

{% highlight js %}
var registrar = NewRegistrar("none","NONE"); // no registrar
var softlayer = NewDnsProvider("softlayer", "SOFTLAYER");

D("example.tld", registrary, DnsProvider(softlayer),
    A("test","1.2.3.4")
);
{%endhighlight%}

## Metadata

This provider does not recognize any special metadata fields unique to SoftLayer dns.
For compatibility with the pre-generated NAMESERVER fields it's recommended to set the NS TTL to 86400 such as:

{% highlight js %}
D("example.tld", registrary, DnsProvider(softlayer),
    {"ns_ttl": "86400"},

    A("test","1.2.3.4")
);
{%endhighlight%}

`ns_ttl` is a standard metadata field that applies to all providers.
