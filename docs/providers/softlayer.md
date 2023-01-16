# SoftLayer DNS Provider

{% hint style="info" %}
**NOTE**: This provider is currently has no maintainer. We are looking for
a volunteer. If this provider breaks it may be disabled or removed if
it can not be easily fixed.
{% endhint %}

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `SOFTLAYER`
along with authentication fields.
Authenticating with SoftLayer requires at least a `username` and `api_key` for authentication. It can also optionally take a `timeout` and `endpoint_url` parameter however these are optional and will use standard defaults if not provided.

Example:

```json
{
  "softlayer": {
    "TYPE": "SOFTLAYER",
    "api_key": "mysecretapikey",
    "username": "myusername"
  }
}
```

To maintain compatibility with existing softlayer CLI services these can also be provided by the `SL_USERNAME` and `SL_API_KEY` environment variables or specified in the `~/.softlayer`, but this is discouraged. More information about these methods can be found at [the softlayer-go library documentation](https://github.com/softlayer/softlayer-go#sessions).

## Usage

An example `dnsconfig.js` configuration:

```javascript
var REG_NONE = NewRegistrar("none"); // no registrar
var DSP_SOFTLAYER = NewDnsProvider("softlayer");

D("example.tld", registrary, DnsProvider(DSP_SOFTLAYER),
    A("test", "1.2.3.4")
);
```

## Metadata
This provider does not recognize any special metadata fields unique to SoftLayer dns.
For compatibility with the pre-generated NAMESERVER fields it's recommended to set the NS TTL to 86400 such as:

```javascript
D("example.tld", REG_NONE, DnsProvider(SOFTLAYER),
    NAMESERVER_TTL(86400),

    A("test", "1.2.3.4")
);
```
