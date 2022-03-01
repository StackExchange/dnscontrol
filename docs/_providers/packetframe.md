---
name: Packetframe
title: Packetframe Provider
layout: default
jsId: PACKETFRAME
---
# Packetframe Provider

## Configuration
In your credentials file, you must provide your Packetframe Token which can be extracted from the `token` cookie on packetframe.com

```json
{
  "packetframe": {
    "token": "your-packetframe-token"
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to Packetframe.

## Usage
Example Javascript:

```js
var REG_NONE = NewRegistrar('none', 'NONE')
var PACKETFRAME = NewDnsProvider("packetframe", "PACKETFRAME");

D("example.tld", REG_NONE, DnsProvider(PACKETFRAME),
    A("test","1.2.3.4")
);
```
