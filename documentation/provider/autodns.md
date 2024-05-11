## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `AUTODNS` along with
[username, password and a context](https://help.internetx.com/display/APIXMLEN/Authentication#Authentication-AuthenticationviaCredentials(username/password/context)).

Example:

{% code title="creds.json" %}
```json
{
  "autodns": {
    "TYPE": "AUTODNS",
    "username": "autodns.service-account@example.com",
    "password": "[***]",
    "context": "33004"
  }
}
```
{% endcode %}

## Usage

An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_AUTODNS = NewDnsProvider("autodns");

D("example.com", REG_NONE, DnsProvider(DSP_AUTODNS),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}
