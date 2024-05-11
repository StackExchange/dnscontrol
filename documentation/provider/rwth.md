## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `RWTH`
along with your [API Token](https://noc-portal.rz.rwth-aachen.de/dns-admin/en/api_tokens).

Example:

{% code title="creds.json" %}
```json
{
  "rwth": {
    "TYPE": "RWTH",
    "api_token": "bQGz0DOi0AkTzG...="
  }
}
```
{% endcode %}

## Metadata
This provider does not recognize any special metadata fields unique to it.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_RWTH = NewDnsProvider("rwth");

D("example.rwth-aachen.de", REG_NONE, DnsProvider(DSP_RWTH),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Caveats
The default TTL is not automatically fetched, as the API does not provide such an endpoint.

The RWTH deploys zones every 15 minutes, so it might take some time for changes to take effect.
