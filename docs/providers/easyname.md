---
name: easyname
title: easyname Provider
---
# easyname Provider

DNSControl's easyname provider supports being a Registrar. Support for being a DNS Provider is not included, but could be added in the future.

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `EASYNAME`
along with [API-Access](https://my.easyname.com/en/account/api) information

Example:

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
An example `dnsconfig.js` configuration:

```javascript
var REG_EASYNAME = NewRegistrar("easyname");

D("example.com", REG_EASYNAME,
  NAMESERVER("ns1.example.com."),
  NAMESERVER("ns2.example.com."),
);
```

## Activation

You must enable API-Access for your account.
