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

## Activation

To obtain the OVH keys, one need to register an app at OVH by following the
[OVH API Getting Started](https://api.ovh.com/g934.first_step_with_api)

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
      "method": "GET",
      "path": "/domain/*"
    },
    {
      "method": "POST",
      "path": "/domain/zone/*"
    },
    {
      "method": "PUT",
      "path": "/domain/zone/*"
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

Open the "validationUrl" in a brower and log in with your OVH account. This will link the app with your account,
authorizing it to access your zones and domains.

Do not forget to fill the `consumer-key` of your `creds.json`.

## New domains

If a domain does not exist in your OVH account, DNSControl
will *not* automatically add it. You'll need to do that via the
control panel manually.

