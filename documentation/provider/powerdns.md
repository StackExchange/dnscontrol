## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `POWERDNS`
along with your [API URL, API Key and Server ID](https://doc.powerdns.com/authoritative/http-api/index.html).
In most cases the Server id is `localhost`.

Example:

{% code title="creds.json" %}
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
{% endcode %}

## Metadata
Following metadata are available:

{% code title="dnsconfig.js" %}
```javascript
{
    'default_ns': [
        'a.example.com.',
        'b.example.com.'
    ],
    'dnssec_on_create': false,
    'zone_kind': 'Native',
}
```
{% endcode %}

- `default_ns` sets the nameserver which are used
- `dnssec_on_create` specifies if DNSSEC should be enabled when creating zones
- `zone_kind` is the type that will be used when creating the zone.
  <br>Can be one of `Native`, `Master` or `Slave`, when not specified it defaults to `Native`.
  <br>Please see [PowerDNS documentation](https://doc.powerdns.com/authoritative/modes-of-operation.html) for explanation of the kinds.
  <br>**Note that these tokens are case-sensitive!**
- `soa_edit_api` is the default SOA serial method that is used for zone created with the API
  <br> Can be one of `DEFAULT`, `INCREASE`, `EPOCH`, `SOA-EDIT` or `SOA-EDIT-INCREASE`, default format is YYYYMMDD01.
  <br>Please see [PowerDNS SOA-EDIT-DNSUPDATE documentation](https://doc.powerdns.com/authoritative/dnsupdate.html#soa-edit-dnsupdate-settings) for explanation of the kinds.
  <br>**Note that these tokens are case-sensitive!**

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_POWERDNS = NewDnsProvider("powerdns");

D("example.com", REG_NONE, DnsProvider(DSP_POWERDNS),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Activation
See the [PowerDNS documentation](https://doc.powerdns.com/authoritative/http-api/index.html) how the API can be enabled.
