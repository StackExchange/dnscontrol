## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `CLOUDNS`
along with your [Api user ID and password](https://www.cloudns.net/wiki/article/42/).

Example:

{% code title="creds.json" %}
```json
{
  "cloudns": {
    "TYPE": "CLOUDNS",
    "auth-id": "12345",
    "sub-auth-id": "12345",
    "auth-password": "your-password"
  }
}
```
{% endcode %}

Current version of provider doesn't support `sub-auth-user`.

## Records

ClouDNS does support DS Record on subdomains (not the apex domain itself).

ClouDNS requires NS records exist for any DS records. No other records for
the same label may exist (A, MX, TXT, etc.). If DNSControl is adding NS and
DS records in the same update, the NS records will be inserted first.

## Metadata
This provider does not recognize any special metadata fields unique to ClouDNS.

## Web Redirects
ClouDNS supports ClouDNS-specific "WR record (web redirects)" for your domains.
Simply use the `CLOUDNS_WR` functions to make redirects like any other record:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_CLOUDNS = NewDnsProvider("cloudns");

D("example.com", REG_NONE, DnsProvider(DSP_CLOUDNS),
  CLOUDNS_WR("@", "http://example.com/"),
  CLOUDNS_WR("www", "http://example.com/"),
END)
```
{% endcode %}

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_CLOUDNS = NewDnsProvider("cloudns");

D("example.com", REG_NONE, DnsProvider(DSP_CLOUDNS),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Activation
[Create Auth ID](https://www.cloudns.net/api-settings/).  Only paid account can use API

## Caveats
ClouDNS does not allow all TTLs, only a specific subset of TTLs. By default, the following [TTLs are supported](https://www.cloudns.net/wiki/article/188/):
- 60  (1 minute)
- 300 (5 minutes)
- 900 (15 minutes)
- 1800 (30 minutes)
- 3600 (1 hour)
- 21600 (6 hours)
- 43200 (12 hours)
- 86400 (1 day)
- 172800 (2 days)
- 259200 (3 days)
- 604800 (1 week)
- 1209600 (2 weeks)
- 2419200 (4 weeks)

The provider will automatically round up your TTL to one of these values. For example, 350 seconds would become 900
seconds, but 300 seconds would stay 300 seconds.
