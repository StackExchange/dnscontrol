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
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Activation
DNSControl depends on a deSEC account auth token.
This token can be obtained by [logging in via the deSEC API](https://desec.readthedocs.io/en/latest/auth/account.html#log-in).

{% hint style="warning" %}
deSEC enforces a daily limit of 300 RRset creation/deletion/modification per
domain. Large changes may have to be done over the course of a few days.  The
integration test suite can not be run in a single session. See
[https://desec.readthedocs.io/en/latest/rate-limits.html#api-request-throttling](https://desec.readthedocs.io/en/latest/rate-limits.html#api-request-throttling)
{% endhint %} 

