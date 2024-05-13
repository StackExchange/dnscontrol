## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `NETCUP`
along with your [api key, password and your customer number](https://www.netcup-wiki.de/wiki/CCP_API#Authentifizierung).

Example:

{% code title="creds.json" %}
```json
{
  "netcup": {
    "TYPE": "NETCUP",
    "api-key": "abc12345",
    "api-password": "abc12345",
    "customer-number": "123456"
  }
}
```
{% endcode %}

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_NETCUP = NewDnsProvider("netcup");

D("example.com", REG_NONE, DnsProvider(DSP_NETCUP),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Caveats
Netcup does not allow any TTLs to be set for individual records. Thus in
the diff/preview it will always show a TTL of 0. `NS` records are also
not currently supported.
