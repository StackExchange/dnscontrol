This is the provider for [Sakura Cloud](https://cloud.sakura.ad.jp/).

## Configuration
To use this provider, add an entry to `creds.json` with `TYPE` set to `SAKURACLOUD`
along with API credentials.

Example:

{% code title="creds.json" %}
```json
{
  "sakuracloud": {
    "TYPE": "SAKURACLOUD",
    "access_token": "your-access-token",
    "access_token_secret": "your-access-token-secret"
  }
}
```
{% endcode %}

The `endpoint` is optional. If omitted, the default endpoint is assumed.

Endpoints are as follows:

* `https://secure.sakura.ad.jp/cloud/zone/is1a/api/cloud/1.1` (Ishikari first Zone)
* `https://secure.sakura.ad.jp/cloud/zone/is1b/api/cloud/1.1` (Ishikari second Zone)
* `https://secure.sakura.ad.jp/cloud/zone/tk1a/api/cloud/1.1` (Tokyo first Zone)
* `https://secure.sakura.ad.jp/cloud/zone/tk1b/api/cloud/1.1` (Tokyo second Zone)

DNS service is independent of zones, so you can use any of these endpoints.
The default is the Ishikari first Zone.

Alternatively you can also use environment variables.

```shell
export SAKURACLOUD_ACCESS_TOKEN="your-access-token"
export SAKURACLOUD_ACCESS_TOKEN_SECRET="your-access-token-secret"
```

{% code title="creds.json" %}
```json
{
  "sakuracloud": {
    "TYPE": "SAKURACLOUD",
    "access_token": "$SAKURACLOUD_ACCESS_TOKEN",
    "access_token_secret": "$SAKURACLOUD_ACCESS_TOKEN_SECRET"
  }
}
```
{% endcode %}

## Metadata
This provider does not recognize any special metadata fields unique to
Sakura Cloud.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_SAKURACLOUD = NewDnsProvider("sakuracloud");

D("example.com", REG_NONE, DnsProvider(DSP_SAKURACLOUD),
  A("test", "192.0.2.1"),
END);
```
{% endcode %}

`NAMESERVER` does not need to be set as the name servers for the
Sakura Cloud provider cannot be changed.

`SOA` cannot be set as SOA record of Sakura Cloud provider cannot be changed.

## Activation
Sakura Cloud depends on an [API Key](https://manual.sakura.ad.jp/cloud/api/apikey.html).

When creating an API key, select "can modify settings" as "Access level".
if you plan to create zones, select "can create and delete resources" as
"Access level".
None of the options in the "Allow access to other services" field need
to be checked.

## Caveats
The limitations of the Sakura Cloud DNS service are described in [the DNS manual](https://manual.sakura.ad.jp/cloud/appliance/dns/index.html), which is written in Japanese.

The limitations not described in that manual are:

* "Null MX", RFC 7505, is not supported.
* SRV records with a Target of "." are not supported.
* SRV records with Port "0" are not supported.
* CAA records with a property value longer than 64 bytes are not allowed.
* Owner names and RDATA targets containing the following labels are not allowed:
    * example
    * exampleN, where N is a numerical character
