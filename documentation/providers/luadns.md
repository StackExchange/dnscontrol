## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `LUADNS`
along with your [email and API key](https://www.luadns.com/api.html#authentication).

Example:

```json
{
  "luadns": {
    "TYPE": "LUADNS",
    "email": "your-email",
    "apikey": "your-api-key"
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to LuaDNS.

## Usage
An example `dnsconfig.js` configuration:

```javascript
var REG_NONE = NewRegistrar("none");
var DSP_LUADNS = NewDnsProvider("luadns");

D("example.tld", REG_NONE, DnsProvider(DSP_LUADNS),
    A("test", "1.2.3.4")
);
```

## Activation
[Create API key](https://api.luadns.com/api_keys).

## Caveats
- LuaDNS cannot change the TTL of the default nameservers.
- This provider does not currently support the "FORWARD" and "REDIRECT" record types.
