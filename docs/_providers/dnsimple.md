---
name: DNSimple
title: DNSimple Provider
layout: default
jsId: DNSIMPLE
---
# DNSimple Provider
## Configuration
In your providers credentials file you must provide a DNSimple account access token:

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
Example Javascript:

```js
var REG_DNSIMPLE = NewRegistrar("dnsimple", "DNSIMPLE");
var DNSIMPLE = NewDnsProvider("dnsimple", "DNSIMPLE");

D("example.tld", REG_DNSIMPLE, DnsProvider(DNSIMPLE),
    A("test","1.2.3.4")
);
```

## Activation
DNSControl depends on a DNSimple account access token.
