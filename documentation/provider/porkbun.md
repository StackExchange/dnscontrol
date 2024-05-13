## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `PORKBUN`
along with your `api_key` and `secret_key`. More info about authentication can be found in [Getting started with the Porkbun API](https://kb.porkbun.com/article/190-getting-started-with-the-porkbun-api).

Example:

{% code title="creds.json" %}
```json
{
  "porkbun": {
    "TYPE": "PORKBUN",
    "api_key": "your-porkbun-api-key",
    "secret_key": "your-porkbun-secret-key",
  }
}
```
{% endcode %}

## Metadata

This provider does not recognize any special metadata fields unique to Porkbun.

## Usage

An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_PORKBUN = NewDnsProvider("porkbun");

D("example.com", REG_NONE, DnsProvider(DSP_PORKBUN),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}
