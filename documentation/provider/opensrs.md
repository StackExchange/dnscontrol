## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `OpenSRS`
along with your OpenSRS credentials.

## Usage

An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_OPENSRS = NewDnsProvider("opensrs");

D("example.com", REG_NONE, DnsProvider(DSP_OPENSRS),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}
