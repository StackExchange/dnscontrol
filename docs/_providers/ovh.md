---
name: Ovh
layout: default
jsId: OVH
---
# OVH Provider

## Configuration

In your providers config json file you must provide a OVH app-key, app-secret-key and consumer-key:

{% highlight json %}
{
  "ovh":{
    "app-key": "your app key",
    "app-secret-key": "your app secret key",
    "consumer-key": "your consumer key"
  }
}
{% endhighlight %}

See [the Activation section](#activation) for details on obtaining these credentials.

## Metadata

This provider does not recognize any special metadata fields unique to OVH.

## Usage

Example javascript:

Example javascript (DNS hosted with OVH):
{% highlight js %}
var REG_OVH = NewRegistrar("ovh", "OVH");
var OVH = NewDnsProvider("ovh", "OVH");

D("example.tld", REG_OVH, DnsProvider(OVH),
    A("test","1.2.3.4")
);
{% endhighlight %}

Example javascript (Registrar only. DNS hosted elsewhere):

{% highlight js %}
var REG_OVH = NewRegistrar("ovh", "OVH");
var R53 = NewDnsProvider("r53", "ROUTE53");

D("example.tld", REG_OVH, DnsProvider(R53),
    A("test","1.2.3.4")
);
{%endhighlight%}


## Activation

To obtain the OVH keys, one need to register an app at OVH by following the
[OVH API Getting Started](https://docs.ovh.com/gb/en/customer/first-steps-with-ovh-api/)

It consist in declaring the app at https://eu.api.ovh.com/createApp/
which gives the `app-key` and `app-secret-key`.

Once done, to obtain the `consumer-key` it is necessary to authorize the just created app
to access the data in a specific account:

{% highlight bash %}
curl -XPOST -H"X-Ovh-Application: <you-app-key>" -H "Content-type: application/json" https://eu.api.ovh.com/1.0/auth/credential -d'{
  "accessRules": [
    {
      "method": "DELETE",
      "path": "/domain/zone/*"
    },
    {
      "method": "GET",
      "path": "/domain/zone/*"
    },
    {
      "method": "POST",
      "path": "/domain/zone/*"
    },
    {
      "method": "PUT",
      "path": "/domain/zone/*"
    },
    {
      "method": "GET",
      "path": "/domain/*"
    },
    {
      "method": "PUT",
      "path": "/domain/*"
    },
    {
      "method": "POST",
      "path": "/domain/*/nameServers/update"
    }
  ]
}'
{% endhighlight %}

It should return something akin to:
{% highlight json %}
{
  "validationUrl": "https://eu.api.ovh.com/auth/?credentialToken=<long-token>",
  "consumerKey": "<your-consumer-key>",
  "state": "pendingValidation"
}
{% endhighlight %}

Open the "validationUrl" in a browser and log in with your OVH account. This will link the app with your account,
authorizing it to access your zones and domains.

Do not forget to fill the `consumer-key` of your `creds.json`.

## New domains

If a domain does not exist in your OVH account, DNSControl
will *not* automatically add it. You'll need to do that via the
control panel manually.

## Dual providers scenario

Since OVH doesn't allow to host DNS for a domain that is not registered in their registrar, some dual providers
scenario are not possible:

| registrar | zone        | working? |
|:---------:|:-----------:|:--------:|
|  OVH      | other       |    √     |
|  OVH      | OVH + other |    √     |
|  other    | OVH         |    X     |

## Caveat

OVH doesn't allow resetting the zone to the OVH DNS through the API. If for any reasons OVH NS entries were
removed the only way to add them back is by using the OVH Control Panel (in the DNS Servers tab, click on the "Reset the
DNS servers" button.
