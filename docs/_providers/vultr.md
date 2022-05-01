---
name: Vultr
title: Vultr Provider
layout: default
jsId: VULTR
---
# Vultr Provider

## Configuration

In your providers config json file you must include a Vultr personal access token:

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

Example javascript:

```js
var VULTR = NewDnsProvider("vultr", "VULTR");

D("example.tld", REG_DNSIMPLE, DnsProvider(VULTR),
    A("test","1.2.3.4")
);
```

## Activation

Vultr depends on a Vultr personal access token.
