## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `GCORE`
along with a Gcore account API token.

Example:

{% code title="creds.json" %}
```json
{
  "gcore": {
    "TYPE": "GCORE",
    "api-key": "your-gcore-api-key"
  }
}
```
{% endcode %}

## Metadata
This provider does not recognize any special metadata fields unique to Gcore.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_GCORE = NewDnsProvider("gcore");

D("example.com", REG_NONE, DnsProvider(DSP_GCORE),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Activation

DNSControl depends on a Gcore account API token.

You can obtain your API token on this page: <https://accounts.gcore.com/profile/api-tokens>
