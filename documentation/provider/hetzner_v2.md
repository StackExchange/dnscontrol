## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `HETZNER_V2`
along with a [Hetzner API Token](https://docs.hetzner.cloud/reference/cloud#getting-started).

Example:

{% code title="creds.json" %}
```json
{
  "hetzner_v2": {
    "TYPE": "HETZNER_V2",
    "api_token": "your-api-token"
  }
}
```
{% endcode %}

## Metadata

This provider does not recognize any special metadata fields unique to Hetzner DNS API.

## Usage

An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_HETZNER = NewDnsProvider("hetzner_v2");

D("example.com", REG_NONE, DnsProvider(DSP_HETZNER),
    A("test", "1.2.3.4"),
);
```
{% endcode %}

## Activation

Create a new API Key in the
[Hetzner Console](https://docs.hetzner.cloud/reference/cloud#getting-started).

## Caveats

### NS

Removing the Hetzner provided NS records at the root is not possible.

### SOA

Hetzner DNS API does not allow changing the SOA record via their API.
There is an alternative method using an import of a full BIND file, but this
 approach does not play nice with incremental changes or ignored records.
At this time you cannot update SOA records via DNSControl.
