# Vultr Provider

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `VULTR`
along with a Vultr personal access token.

Example:

```json
{
  "vultr": {
    "TYPE": "VULTR",
    "token": "your-vultr-personal-access-token"
  }
}
```

## Metadata

This provider does not recognize any special metadata fields unique to Vultr.

## Usage

An example `dnsconfig.js` configuration:

```javascript
var DSP_VULTR = NewDnsProvider("vultr");

D("example.tld", REG_DNSIMPLE, DnsProvider(DSP_VULTR),
    A("test", "1.2.3.4")
);
```

## Activation

Vultr depends on a Vultr personal access token.
