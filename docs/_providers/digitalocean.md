---
name: DigitalOcean
title: DigitalOcean Provider
layout: default
jsId: DIGITALOCEAN
---
# DigitalOcean Provider

## Configuration
In your credentials file, you must provide your
[DigitalOcean OAuth Token](https://cloud.digitalocean.com/settings/applications)

{% highlight json %}
{
  "digitalocean": {
    "token": "your-digitalocean-ouath-token"
  }
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to DigitalOcean.

## Usage
Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none', 'NONE')
var DIGITALOCEAN = NewDnsProvider("digitalocean", "DIGITALOCEAN");

D("example.tld", REG_NONE, DnsProvider(DIGITALOCEAN),
    A("test","1.2.3.4")
);
{%endhighlight%}

## Activation
[Create OAuth Token](https://cloud.digitalocean.com/settings/applications)

## Limitations

- Digialocean DNS doesn't support `;` value with CAA-records ([DigitalOcean documentation](https://www.digitalocean.com/docs/networking/dns/how-to/create-caa-records/))
- No support for TXT records with multiple strings, as the API prevents espacing quotes.
