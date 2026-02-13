## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `POWERDNS`
along with your [API URL, API Key and Server ID](https://doc.powerdns.com/authoritative/http-api/index.html).
In most cases the Server id (`serverName`) is `localhost`.

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
Following provider metadata are available:

{% code title="dnsconfig.js" %}
```javascript
var DSP_POWERDNS = NewDnsProvider("pdns", {
    "default_ns": [
        "a.example.com.",
        "b.example.com."
    ],
    "dnssec_on_create": false,
    "zone_kind": "Native",
    "use_views": true
});
```
{% endcode %}

- `default_ns` sets the nameservers which are used.
- `dnssec_on_create` specifies if DNSSEC should be enabled when creating zones.
- `zone_kind` is the type that will be used when creating the zone.
  <br>Can be one of `Native`, `Master` or `Slave`, when not specified it defaults to `Native`.
  <br>Please see [PowerDNS documentation](https://doc.powerdns.com/authoritative/modes-of-operation.html) for explanation of the kinds.
  <br>**Note that these tokens are case-sensitive!**
- `soa_edit_api` is the default SOA serial method that is used for zone created with the API
  <br> Can be one of `DEFAULT`, `INCREASE`, `EPOCH`, `SOA-EDIT` or `SOA-EDIT-INCREASE`, default format is YYYYMMDD01.
  <br>Please see [PowerDNS SOA-EDIT-DNSUPDATE documentation](https://doc.powerdns.com/authoritative/dnsupdate.html#soa-edit-dnsupdate-settings) for explanation of the kinds.
  <br>**Note that these tokens are case-sensitive!**
- `use_views` enables mapping dnscontrol tags to PowerDNS views.
  <br>Set to `true` to enable, defaults to `false`.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_POWERDNS = NewDnsProvider("powerdns");

D("example.com", REG_NONE, DnsProvider(DSP_POWERDNS),
    A("test", "1.2.3.4"),
);
```
{% endcode %}

## Activation
See the [PowerDNS documentation](https://doc.powerdns.com/authoritative/http-api/index.html) how the API can be enabled.

## Tags and Variants
If you use a dnscontrol *tag* (like `example.com!internal`) it will be mapped to a powerdns *variant* (like `example.com..internal`) when `use_views` is enabled in the provider metadata.

See [PowerDNS documentation on Views](https://doc.powerdns.com/authoritative/views.html) for details on how to setup networks and views for these variants.

## Caveats

### SOA Records
The SOA record is supported for use, but behavior is slightly different than expected.
If the SOA record is used, [PowerDNS will not increase the serial](https://doc.powerdns.com/authoritative/dnsupdate.html#soa-serial-updates) if the SOA record content changes.
This itself comes with exceptions as well, if the `SOA-EDIT-API` is changed to a different value the logic will update the serial to a new value.
See [this issue for detailed testing](https://github.com/StackExchange/dnscontrol/pull/3404#issuecomment-2628989200) of behavior.

The recommended procedure when changing the SOA record contents is to update the SOA record alone.
Updates to other records will be done if changes are present, but the serial **will not change**. The serial will update once a new push is done that does not include an SOA record change.
