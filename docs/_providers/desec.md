---
name: deSEC
title: deSEC Provider
layout: default
jsId: DESEC
---
# deSEC Provider
## Configuration
In your providers credentials file you must provide a deSEC account auth token:

{% highlight json %}
{
  "desec": {
    "auth-token": "your-deSEC-auth-token"
  }
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to deSEC.

## Usage
Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none', 'NONE');    // No registrar.
var deSEC = NewDnsProvider('desec', 'DESEC');  // deSEC

D('example.tld', REG_NONE, DnsProvider(deSEC),
    A('test','1.2.3.4')
);
{% endhighlight %}

## Activation
DNSControl depends on a deSEC account auth token.
This token can be obtained by logging in via the deSEC API: https://desec.readthedocs.io/en/latest/auth/account.html#log-in