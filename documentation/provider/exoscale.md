## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `EXOSCALE`
along with your Exoscale credentials.

## Usage

An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_EXOSCALE = NewDnsProvider("exoscale");

D("example.com", REG_NONE, DnsProvider(DSP_EXOSCALE),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}
