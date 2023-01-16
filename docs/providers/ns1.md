# NS1 Provider

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `NS1`
along with your NS1 api key.

Example:

```json
{
  "ns1": {
    "TYPE": "NS1",
    "api_token": "your-ns1-token"
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to NS1.

## Usage
An example `dnsconfig.js` configuration:

```javascript
var REG_NONE = NewRegistrar("none");
var DSP_NS1 = NewDnsProvider("ns1");

D("example.tld", REG_NONE, DnsProvider(DSP_NS1),
    A("test", "1.2.3.4")
);
```

