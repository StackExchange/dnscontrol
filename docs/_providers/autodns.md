---
name: AutoDNS
title: AutoDNS (InternetX)
layout: default
jsId: AUTODNS
---

# AutoDNS Provider

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `AUTODNS` along with
[username, password and a context](https://help.internetx.com/display/APIXMLEN/Authentication#Authentication-AuthenticationviaCredentials(username/password/context)).

Example:

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
var REG_NONE = NewRegistrar("none");
var AUTODNS = NewDnsProvider("autodns");

D("example.tld", REG_NONE, DnsProvider(AUTODNS),
    A("test","1.2.3.4")
);
{%endhighlight%}
