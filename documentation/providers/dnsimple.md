## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `DNSIMPLE`
along with a DNSimple account access token.

You can also set the `baseurl` to use [DNSimple's free sandbox](https://developer.dnsimple.com/sandbox/) for testing.

Examples:

{% code title="creds.json" %}
```json
{
  "dnsimple": {
    "TYPE": "DNSIMPLE",
    "token": "your-dnsimple-account-access-token"
  },
  "dnsimple_sandbox": {
    "TYPE": "DNSIMPLE",
    "baseurl": "https://api.sandbox.dnsimple.com",
    "token": "your-sandbox-account-access-token"
  }
}
```
{% endcode %}

## Metadata
This provider does not recognize any special metadata fields unique to DNSimple.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_DNSIMPLE = NewRegistrar("dnsimple");
var DSP_DNSIMPLE = NewDnsProvider("dnsimple");

D("example.com", REG_DNSIMPLE, DnsProvider(DSP_DNSIMPLE),
    A("test", "1.2.3.4")
);
```
{% endcode %}

## Activation
DNSControl depends on a DNSimple account access token.

## Caveats

None at this time
