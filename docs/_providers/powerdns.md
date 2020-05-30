---
name: PowerDNS
title: PowerDNS Provider
layout: default
jsId: POWERDNS
---
# PowerDNS Provider

## Configuration
In your credentials file, you must provide your [API URL, API Key and Server ID](https://doc.powerdns.com/authoritative/http-api/index.html). 

In most cases the Server id is `localhost`

{% highlight json %}
{
  "powerdns": {
    "apiurl": "http://localhost",
    "apikey": "your-key",
    "servername": "localhost"
  }
}
{% endhighlight %}

## Metadata
Following metadata are available:

{% highlight js %}
{
    'default_ns': [
        'a.example.com.',
        'b.example.com.'
    ],
    'dnssec_on_create': false
}
{% endhighlight %}

- `default_ns` sets the nameserver which are used
- `dnssec_on_create` specifies if DNSSEC should be enabled when creating zones

## Usage
Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none', 'NONE')
var POWERDNS = NewDnsProvider("powerdns", "POWERDNS");

D("example.tld", REG_NONE, DnsProvider(POWERDNS),
    A("test","1.2.3.4")
);
{%endhighlight%}

## Activation
See the [PowerDNS documentation](https://doc.powerdns.com/authoritative/http-api/index.html) how the API can be enabled.

## Caveats
Currently it is only possible to enable DNSSec while creating a new zone.
On-demand activation/deactivation of DNSSec will be added later.