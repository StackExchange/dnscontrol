This is the provider for [AdGuardHome](https://github.com/AdguardTeam/AdGuardHome).

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
    "host": "foo.com"
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
    A_PASSTHROUGH("abc", ""),
    AAAA_PASSTHROUGH("abc", ""),
);
```
{% endcode %}

## Integration testing

The integration tests assume that Cloudflare Workers are enabled and the credentials used
have the required permissions listed above.  The flag `-cfworkers=false` will disable tests related to Workers.
This flag is intended for use with legacy domains where the integration test credentials do not
have access to read/edit Workers. This flag will eventually go away.

```shell
cd integrationTest              # NOTE: Not needed if already in that subdirectory
go test -v -verbose -profile CLOUDFLAREAPI -cfworkers=false
```

When `-cfworkers=false` is set, tests related to Workers are skipped.  The Account ID is not required.
