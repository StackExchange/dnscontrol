This is the provider for [Infomaniak](https://www.infomaniak.com/).

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `INFOMANIAK` along with a Infomaniak account personal access token.

Examples:

{% code title="creds.json" %}
```json
{
  "infomaniak": {
    "TYPE": "INFOMANIAK",
    "token": "your-infomaniak-account-access-token",
  }
}
```
{% endcode %}

## Metadata
This provider does not recognize any special metadata fields unique to Infomaniak.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_INFOMANIAK = NewDnsProvider("infomaniak");

D("example.com", REG_NONE, DnsProvider(DSP_INFOMANIAK),
    A("test", "1.2.3.4"),
);
```
{% endcode %}

## Activation
DNSControl depends on a Infomaniak account personal access token.
