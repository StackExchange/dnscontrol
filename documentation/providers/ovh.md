## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `OVH`
along with a OVH app-key, app-secret-key, consumer-key and optionally endpoint.

Example:

{% code title="creds.json" %}
```json
{
  "ovh": {
    "TYPE": "OVH",
    "app-key": "your app key",
    "app-secret-key": "your app secret key",
    "consumer-key": "your consumer key",
    "endpoint": "eu"
  }
}
```
{% endcode %}

See [the Activation section](#activation) for details on obtaining these credentials.

`endpoint` can take the following values:

* `eu` (the default), for connecting to the OVH European endpoint
* `ca` for connecting to OVH Canada API endpoint
* `us` for connecting to the OVH USA API endpoint
* an url for connecting to a different endpoint than the ones above

## Metadata

This provider does not recognize any special metadata fields unique to OVH.

## Usage

An example configuration: (DNS hosted with OVH):

{% code title="dnsconfig.js" %}
```javascript
var REG_OVH = NewRegistrar("ovh");
var DSP_OVH = NewDnsProvider("ovh");

D("example.com", REG_OVH, DnsProvider(DSP_OVH),
    A("test", "1.2.3.4")
);
```
{% endcode %}

An example configuration: (Registrar only. DNS hosted elsewhere)

{% code title="dnsconfig.js" %}
```javascript
var REG_OVH = NewRegistrar("ovh");
var DSP_R53 = NewDnsProvider("r53");

D("example.com", REG_OVH, DnsProvider(DSP_R53),
    A("test", "1.2.3.4")
);
```
{% endcode %}

## Activation

To obtain the OVH keys, one need to register an app at OVH by following the
[OVH API Getting Started](https://docs.ovh.com/gb/en/customer/first-steps-with-ovh-api/)

It consist in declaring the app at https://eu.api.ovh.com/createApp/
which gives the `app-key` and `app-secret-key`. If your domains and zones are located in another region, see below for the correct url to use.

Once done, to obtain the `consumer-key` it is necessary to authorize the just created app
to access the data in a specific account:

```shell
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
```

It should return something akin to:

{% code title="creds.json" %}
```json
{
  "validationUrl": "https://eu.api.ovh.com/auth/?credentialToken=<long-token>",
  "consumerKey": "<your-consumer-key>",
  "state": "pendingValidation"
}
```
{% endcode %}

Open the "validationUrl" in a browser and log in with your OVH account. This will link the app with your account,
authorizing it to access your zones and domains.

Do not forget to fill the `consumer-key` of your `creds.json`.

For accessing the other international endpoints such as US and CA, change the `https://eu.api.ovh.com` used above to one of the following:

* Canada endpoint: `https://ca.api.ovh.com`
* US endpoint: `https://api.us.ovhcloud.com`

Do not forget to fill the `endpoint` of your `creds.json` if you use an endpoint different than the EU one.

## New domains

If a domain does not exist in your OVH account, DNSControl
will *not* automatically add it. You'll need to do that via the
control panel manually.

## Dual providers scenario

OVH now allows to host DNS zone for a domain that is not registered in their registrar (see: https://www.ovh.com/manager/web/#/zone). The following dual providers scenario are supported:

| registrar | zone        | working? |
|:---------:|:-----------:|:--------:|
|  OVH      | other       |    √     |
|  OVH      | OVH + other |    √     |
|  other    | OVH         |    √     |

## Caveat

OVH doesn't allow resetting the zone to the OVH DNS through the API. If for any reasons OVH NS entries were
removed the only way to add them back is by using the OVH Control Panel (in the DNS Servers tab, click on the "Reset the
DNS servers" button.
