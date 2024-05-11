## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `DNSMADEEASY`
along with your `api_key` and `secret_key`. More info about authentication can be found in [DNS Made Easy API docs](https://api-docs.dnsmadeeasy.com/).

Example:

{% code title="creds.json" %}
```json
{
  "dnsmadeeasy": {
    "TYPE": "DNSMADEEASY",
    "api_key": "1c1a3c91-4770-4ce7-96f4-54c0eb0e457a",
    "secret_key": "e2268cde-2ccd-4668-a518-8aa8757a65a0"
  }
}
```
{% endcode %}

## Records

ALIAS/ANAME records are supported.

This provider does not support HTTPRED records.

SPF records are ignored by this provider. Use TXT records instead.

## Metadata
This provider does not recognize any special metadata fields unique to DNS Made Easy.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_DNSMADEEASY = NewDnsProvider("dnsmadeeasy");

D("example.com", REG_NONE, DnsProvider(DSP_DNSMADEEASY),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Activation
You can generate your `api_key` and `secret_key` in [Control Panel](https://cp.dnsmadeeasy.com/) in Account Information in Config menu.

API is only available for Business plan and higher plans.

## Caveats

### Global Traffic Director
Global Traffic Director feature is not supported.

## Development

### Debugging
Set `DNSMADEEASY_DEBUG_HTTP` environment variable to dump all API calls made by this provider.

### Testing
Set `sandbox` key to any non-empty value in credentials JSON alongside `api_key` and `secret_key` to make all API calls against DNS Made Easy sandbox environment.
