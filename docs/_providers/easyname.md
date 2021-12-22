---
name: easyname
title: easyname Provider
layout: default
jsId: EASYNAME
---
# easyname Provider

DNSControl's easyname provider supports being a Registrar. Support for being a DNS Provider is not included, but could be added in the future.

## Configuration
In your credentials file, you must provide your [API-Access](https://my.easyname.com/en/account/api) information

{% highlight json %}
{
  "easyname": {
    "userid": 12345,
    "email": "example@example.com",
    "apikey": "API Key",
    "authsalt": "API Authentication Salt",
    "signsalt": "API Signing Salt"
  }
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to easyname.

## Usage
Example Javascript:

{% highlight js %}
var REG_EASYNAME = NewRegistrar('easyname', 'EASYNAME');

D("example.com", REG_EASYNAME,
  NAMESERVER("ns1.example.com."),
  NAMESERVER("ns2.example.com."),
);
{% endhighlight %}

## Activation

You must enable API-Access for your account.
