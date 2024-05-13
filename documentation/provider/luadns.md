## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `LUADNS`
along with your [email and API key](https://www.luadns.com/api.html#authentication).

Example:

{% code title="creds.json" %}
```json
{
  "luadns": {
    "TYPE": "LUADNS",
    "email": "your-email",
    "apikey": "your-api-key"
  }
}
```
{% endcode %}

## Metadata
This provider does not recognize any special metadata fields unique to LuaDNS.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_LUADNS = NewDnsProvider("luadns");

D("example.com", REG_NONE, DnsProvider(DSP_LUADNS),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Activation
[Create API key](https://api.luadns.com/api_keys).

## Caveats
- LuaDNS cannot change the default nameserver TTL in `nameserver_ttl`, it is forced to fixed at 86400("1d").
This is not the case if you are using vanity nameservers.
- This provider does not currently support the "FORWARD" and "REDIRECT" record types.
- The API is available on the LuaDNS free plan, but due to the limit of 30 records, some tests will fail when doing integration tests.
