---
name: DNSimple
title: DNSimple Provider
layout: default
jsId: DNSIMPLE
---
# DNSimple Provider
## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `DIGITALOCEAN`
along with a DNSimple account access token.

Example:

```json
{
  "dnsimple": {
    "TYPE": "DNSIMPLE",
    "token": "your-dnsimple-account-access-token"
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to DNSimple.

## Usage
An example `dnsconfig.js` configuration:

```js
var REG_DNSIMPLE = NewRegistrar("dnsimple");
var DSP_DNSIMPLE = NewDnsProvider("dnsimple");

D("example.tld", REG_DNSIMPLE, DnsProvider(DSP_DNSIMPLE),
    A("test", "1.2.3.4")
);
```

## Activation
DNSControl depends on a DNSimple account access token.
