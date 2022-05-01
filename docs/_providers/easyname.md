---
name: easyname
title: easyname Provider
layout: default
jsId: EASYNAME
---
# easyname Provider

DNSControl's easyname provider supports being a Registrar. Support for being a DNS Provider is not included, but could be added in the future.

## Configuration
In your credentials file, you must provide your [API-Access](https://my.easyname.com/en/account/api) information

```json
{
  "easyname": {
    "TYPE": "EASYNAME",
    "apikey": "API Key",
    "authsalt": "API Authentication Salt",
    "email": "example@example.com",
    "signsalt": "API Signing Salt",
    "userid": 12345
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to easyname.

## Usage
Example Javascript:

```js
var REG_EASYNAME = NewRegistrar('easyname', 'EASYNAME');

D("example.com", REG_EASYNAME,
  NAMESERVER("ns1.example.com."),
  NAMESERVER("ns2.example.com."),
);
```

## Activation

You must enable API-Access for your account.
