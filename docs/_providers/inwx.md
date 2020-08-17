---
name: INWX
layout: default
jsId: INWX
---

# INWX

INWX.de is a Berlin-based domain registrar.

## Configuration
In your `creds.json` file you must provide your INWX
username and password:

{% highlight json %}
{
  "inwx":{
    "username": "yourUsername",
    "password": "yourPassword"
  }
}
{% endhighlight %}

### Two factor authentication

If two factor authentication has been enabled you will also need to provide a valid TOTP number.
This can also be done
via an environment variable:
{% highlight json %}
{
  "inwx":{
    "username": "yourUsername",
    "password": "yourPassword",
    "totp": "$INWX_TOTP"
  }
}
{% endhighlight %}

and then you can run

{% highlight bash %}
$ INWX_TOTP=12345 dnscontrol preview
{% endhighlight %}

It is also possible to directly provide the shared TOTP secret using the key "totp-key" in `creds.json`.
This secret is only shown once when two factor authentication is enabled and you'll have to make sure to write it down then. 

**Important Notes**:
* Anyone with access to this `creds.json` file will have *full* access to your INWX account and will be able to transfer and/or delete your domains
* Storing the shared secret together with the password weakens two factor authentication because both factors are stored in a single place.

{% highlight json %}
{
  "inwx":{
    "username": "yourUsername",
    "password": "yourPassword",
    "totp-key": "yourTOTPSharedSecret"
  }
}
{% endhighlight %}


### Sandbox
You can optionally also specify sandbox with a value of 1 to
redirect all requests to the sandbox API instead:
{% highlight json %}
{
  "inwx":{
    "username": "yourUsername",
    "password": "yourPassword",
    "sandbox": "1"
  }
}
{% endhighlight %}

If sandbox is omitted or set to any other value the production
API will be used.


## Metadata
This provider does not recognize any special metadata fields unique to
INWX.

## Usage
Example Javascript for `example.tld` registered with INWX
and delegated to CloudFlare:

{% highlight js %}
var regInwx = NewRegistrar('inwx', 'INWX')
var dnsCF = NewDnsProvider('cloudflare', 'CLOUDFLAREAPI')

D("example.tld", regInwx, DnsProvider(dnsCF),
    A("test","1.2.3.4")
);

{%endhighlight%}



