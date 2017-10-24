---
name: NS1
title: NS1 Provider
layout: default
jsId: NS1
---
# NS1 Provider

## Configuration

In your credentials json file you must provide your NS1 api key:

{% highlight json %}
{
  "ns1":{
    "api_token": "your-ns1-token"
  }
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to NS1.

## Usage
Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none', 'NONE')
var NS1 = NewDnsProvider("ns1", "NS1");

D("example.tld", REG_NONE, DnsProvider(NS1),
    A("test","1.2.3.4")
);
{% endhighlight %}

