---
name: AutoDNS
title: AutoDNS (InternetX)
layout: default
jsId: AUTODNS
---

# AutoDNS Provider

## Configuration

In your credentials file, you must provide [username, password and a context](https://help.internetx.com/display/APIXMLEN/Authentication#Authentication-AuthenticationviaCredentials(username/password/context)).

{% highlight json %}
{
  "autodns": {
    "TYPE": "AUTODNS",
    "username": "autodns.service-account@example.com",
    "password": "[***]",
    "context": "33004"
  }
}
{% endhighlight %}

## Usage

Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none', 'NONE');
var AUTODNS = NewDnsProvider("autodns", "AUTODNS");

D("example.tld", REG_NONE, DnsProvider(AUTODNS),
    A("test","1.2.3.4")
);
{%endhighlight%}
