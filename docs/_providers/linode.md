---
name: Linode
title: Linode Provider
layout: default
jsId: LINODE
---
# Linode Provider

## Configuration
In your credentials file, you must provide your
[Linode Personal Access Token](https://cloud.linode.com/profile/tokens)

{% highlight json %}
{
  "linode": {
    "token": "your-linode-personal-access-token"
  }
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to Linode.

## Usage
Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none', 'NONE')
var LINODE = NewDnsProvider("linode", "LINODE");

D("example.tld", REG_NONE, DnsProvider(LINODE),
    A("test","1.2.3.4")
);
{%endhighlight%}

## Activation
[Create Personal Access Token](https://cloud.linode.com/profile/tokens)
