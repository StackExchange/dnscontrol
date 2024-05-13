## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `LINODE`
along with your [Linode Personal Access Token](https://cloud.linode.com/profile/tokens).

Example:

{% code title="creds.json" %}
```json
{
  "linode": {
    "TYPE": "LINODE",
    "token": "your-linode-personal-access-token"
  }
}
```
{% endcode %}

## Metadata
This provider does not recognize any special metadata fields unique to Linode.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_LINODE = NewDnsProvider("linode");

D("example.com", REG_NONE, DnsProvider(DSP_LINODE),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Activation
[Create Personal Access Token](https://cloud.linode.com/profile/tokens)

## Caveats
Linode does not allow all TTLs, but only a specific subset of TTLs. The following TTLs are supported
([source](https://www.linode.com/docs/api/domains/#domains-list__responses)):

- 0 (Default, currently equivalent to 1209600, or 14 days)
- 300
- 3600
- 7200
- 14400
- 28800
- 57600
- 86400
- 172800
- 345600
- 604800
- 1209600
- 2419200

The provider will automatically round up your TTL to one of these values. For example, 600 seconds would become 3600
seconds, but 300 seconds would stay 300 seconds.
