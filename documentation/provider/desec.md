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
);
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

Upon domain creation, the DNSKEY and DS records needed for DNSSEC setup are
printed in the command output. If you need these values later, get them from
the deSEC web interface or query deSEC nameservers for the CDS records. For
example: `dig +short @ns1.desec.io example.com CDS` will return the published
CDS records which can be used to insert the required DS records into the parent
zone.