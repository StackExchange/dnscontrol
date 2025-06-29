This is the provider for [AdGuardHome](https://github.com/AdguardTeam/AdGuardHome).

## Important notes

This provider only supports the following record types.

* A
* AAAA
* CNAME
* ALIAS
* ADGUARDHOME_A_PASSTHROUGH
* ADGUARDHOME_AAAA_PASSTHROUGH

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `ADGUARDHOME`.

Required fields include:

* `username` and `password`: Authentication information
* `host`: The hostname/address of AdGuard Home instance

Example:

{% code title="creds.json" %}
```json
{
  "adguard_home": {
    "TYPE": "ADGUARDHOME",
    "username": "admin",
    "password": "your-password",
    "host": "https://foo.com"
  }
}
```
{% endcode %}

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_ADGUARDHOME = NewDnsProvider("adguard_home");

// Example domain where the CF proxy abides by the default (off).
D("example.com", REG_NONE, DnsProvider(DSP_ADGUARDHOME),
    A("foo", "1.2.3.4"),
    AAAA("another", "2003::1"),
    ALIAS("@", "www.example.com."),
    CNAME("myalias", "www.example.com."),
    ADGUARDHOME_A_PASSTHROUGH("abc", ""),
    ADGUARDHOME_AAAA_PASSTHROUGH("abc", ""),
);
```
{% endcode %}

