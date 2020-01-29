---
name: Internet.bs
title: Internet.bs Provider
layout: default
jsId: INTERNETBS
---
# Internet.bs Provider

DNSControl's Internet.bs provider supports being a Registrar. Support for being a DNS Provider is not included, but could be added in the future.

## Configuration
In your credentials file, you must provide your API key and account password 

{% highlight json %}
{
  "internetbs": {
    "api-key": "your-api-key",
    "password": "account-password"
  }
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to Internet.bs.

## Usage
Example Javascript:

{% highlight js %}
var REG_INTERNETBS = NewRegistrar('internetbs', 'INTERNETBS');

D("example.com", REG_INTERNETBS,
  NNAMESERVER("ns1.example.com."),
  NNAMESERVER("ns2.example.com."),
);
{% endhighlight %}

## Activation

Pay attention, you need to define white list of IP for API. But you always can change it on `My Profile > Reseller Settings`   
