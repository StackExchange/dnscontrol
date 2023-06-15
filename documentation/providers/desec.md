## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `DESEC`
along with a deSEC account auth token.

Example:

{% code title="creds.json" %}
```json
{
  "desec": {
    "TYPE": "DESEC",
    "auth-token": "your-deSEC-auth-token"
  }
}
```
{% endcode %}

## Metadata
This provider does not recognize any special metadata fields unique to deSEC.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_DESEC = NewDnsProvider("desec");

D("example.com", REG_NONE, DnsProvider(DSP_DESEC),
    A("test", "1.2.3.4")
);
```
{% endcode %}

## Activation
DNSControl depends on a deSEC account auth token.
This token can be obtained by [logging in via the deSEC API](https://desec.readthedocs.io/en/latest/auth/account.html#log-in).
