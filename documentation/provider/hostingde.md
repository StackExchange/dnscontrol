## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `HOSTINGDE`
along with your [`authToken` and optionally an `ownerAccountId`](https://www.hosting.de/api/#requests-and-authentication).

Example:

{% code title="creds.json" %}
```json
{
  "hosting.de": {
    "TYPE": "HOSTINGDE",
    "authToken": "YOUR_API_KEY"
  }
}
```
{% endcode %}

## Usage

An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_HOSTINGDE = NewRegistrar("hosting.de");
var DSP_HOSTINGDE = NewDnsProvider("hosting.de");

D("example.com", REG_HOSTINGDE, DnsProvider(DSP_HOSTINGDE),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Using this provider with http.net and others

http.net and other DNS service providers use an API that is compatible with hosting.de's API.
Using them requires setting the `baseURL` and (optionally) overriding the default nameservers.

### Example http.net configuration

An example `creds.json` configuration:

{% code title="creds.json" %}
```json
{
  "http.net": {
    "TYPE": "HOSTINGDE",
    "authToken": "YOUR_API_KEY",
    "baseURL": "https://partner.http.net"
  }
}
```
{% endcode %}

An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_HTTPNET = NewRegistrar("http.net");

var DSP_HTTPNET = NewDnsProvider("http.net", {
  "default_ns": [
    "ns1.routing.net.",
    "ns2.routing.net.",
    "ns3.routing.net.",
  ],
});
```
{% endcode %}

#### Why this works

hosting.de has the concept of _nameserver sets_ but this provider does not implement it.
The `HOSTINGDE` provider **ignores the default nameserver set** defined in your account to avoid unintentional changes and consolidate the full configuration in DNSControl.
Instead, it uses hosting.de's nameservers (`ns1.hosting.de.`, `ns2.hosting.de.`, and `ns3.hosting.de.`) by default, regardless of your account settings.
Using the `default_ns` metadata, the default nameserver set can be overwritten.
