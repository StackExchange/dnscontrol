---
name: NS1
layout: default
jsId: NS1
---
# NS1 Provider

NS1 provides a dns provider implementation for ns1 dns.

## Configuration

In your providers config json file you must provide your ns1 api key:

{% highlight json %}
{
  "ns1":{
    "api_token": "your-ns1-token"
  }
}
{% endhighlight %}

## Metadata

This provider does not recognize any special metadata fields unique to ns1.

## Usage

Example javascript:

{% highlight js %}
var NS1 = NewDnsProvider("ns1", "NS1");

D("example.tld", MY_REGISTRAR, DnsProvider(NS1),
    A("test","1.2.3.4")
);
{% endhighlight %}

