---
name: Hurricane Electric DNS
title: Hurricane Electric DNS Provider
layout: default
jsId: HEDNS
---
# Hurricane Electric DNS Provider
## Configuration
In your `creds.json` file you must provide your `dns.he.net` account username and password, along wit

In your `creds.json` file you must provide your INWX
username and password:

{% highlight json %}
{
  "hedns":{
    "username": "yourUsername",
    "password": "yourPassword"
  }
}
{% endhighlight %}

### Two factor authentication

If two-factor authentication has been enabled on your account you will also need to provide a valid TOTP code.
This can also be done via an environment variable:

{% highlight json %}
{
  "hedns":{
    "username": "yourUsername",
    "password": "yourPassword",
    "totp": "$HEDNS_TOTP"
  }
}
{% endhighlight %}

and then you can run

{% highlight bash %}
$ HEDNS_TOTP=12345 dnscontrol preview
{% endhighlight %}

It is also possible to directly provide the shared TOTP secret using the key "totp-key" in `creds.json`. This secret is
only available when first enabling two-factor authentication.

**Important Notes**:
* Anyone with access to this `creds.json` file will have *full* access to your Hurrican Electric account and will be 
  able to modify and delete your DNS entries
* Storing the shared secret together with the password weakens two factor authentication because both factors are stored
  in a single place.

{% highlight json %}
{
  "hedns":{
    "username": "yourUsername",
    "password": "yourPassword",
    "totp-key": "yourTOTPSharedSecret"
  }
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to Hurricane Electric DNS.

## Usage
Example Javascript:

{% highlight js %}
var DNSIMPLE = NewDnsProvider("hedns", "HEDNS");

D("example.tld", REG_DNSIMPLE, DnsProvider(HEDNS),
    A("test","1.2.3.4")
);
{% endhighlight %}
