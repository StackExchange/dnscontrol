---
name: Packetframe
---
# Packetframe Provider

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `PACKETFRAME`
along with your Packetframe Token which can be extracted from the `token` cookie on packetframe.com

Example:

```json
{
  "packetframe": {
    "TYPE": "PACKETFRAME",
    "token": "your-packetframe-token"
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to Packetframe.

## Usage
An example `dnsconfig.js` configuration:

```javascript
var REG_NONE = NewRegistrar("none");
var DSP_PACKETFRAME = NewDnsProvider("packetframe");

D("example.tld", REG_NONE, DnsProvider(DSP_PACKETFRAME),
    A("test", "1.2.3.4")
);
```
