# PowerDNS Provider

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `POWERDNS`
along with your [API URL, API Key and Server ID](https://doc.powerdns.com/authoritative/http-api/index.html).
In most cases the Server id is `localhost`.

Example:

```json
{
  "powerdns": {
    "TYPE": "POWERDNS",
    "apiKey": "your-key",
    "apiUrl": "http://localhost",
    "serverName": "localhost"
  }
}
```

## Metadata
Following metadata are available:

```javascript
{
    'default_ns': [
        'a.example.com.',
        'b.example.com.'
    ],
    'dnssec_on_create': false
}
```

- `default_ns` sets the nameserver which are used
- `dnssec_on_create` specifies if DNSSEC should be enabled when creating zones

## Usage
An example `dnsconfig.js` configuration:

```javascript
var REG_NONE = NewRegistrar("none");
var DSP_POWERDNS = NewDnsProvider("powerdns");

D("example.tld", REG_NONE, DnsProvider(DSP_POWERDNS),
    A("test", "1.2.3.4")
);
```

## Activation
See the [PowerDNS documentation](https://doc.powerdns.com/authoritative/http-api/index.html) how the API can be enabled.
