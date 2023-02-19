## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `AUTODNS` along with
[username, password and a context](https://help.internetx.com/display/APIXMLEN/Authentication#Authentication-AuthenticationviaCredentials(username/password/context)).

Example:

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

## Usage

An example `dnsconfig.js` configuration:

```javascript
var REG_NONE = NewRegistrar("none");
var DSP_AUTODNS = NewDnsProvider("autodns");

D("example.tld", REG_NONE, DnsProvider(DSP_AUTODNS),
    A("test", "1.2.3.4")
);
```
